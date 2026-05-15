package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ResumeParser struct {
	pythonPath string
}

type ParseResumeResponse struct {
	FileName string `json:"fileName"`
	Text     string `json:"text"`
}

func NewResumeParser(pythonPath string) *ResumeParser {
	return &ResumeParser{
		pythonPath: strings.TrimSpace(pythonPath),
	}
}

func (p *ResumeParser) Parse(ctx context.Context, fileName string, fileBytes []byte) (ParseResumeResponse, error) {
	name := strings.TrimSpace(fileName)
	if name == "" {
		return ParseResumeResponse{}, errors.New("fileName is required")
	}
	if len(fileBytes) == 0 {
		return ParseResumeResponse{}, errors.New("resume file is empty")
	}
	if p.pythonPath == "" {
		return ParseResumeResponse{}, errors.New("python runtime is not configured")
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".md", ".json":
		return ParseResumeResponse{
			FileName: name,
			Text:     strings.TrimSpace(string(fileBytes)),
		}, nil
	case ".pdf", ".docx":
	default:
		return ParseResumeResponse{}, errors.New("unsupported resume format, only pdf/docx/txt/md/json are allowed")
	}

	payload := map[string]string{
		"fileName": name,
		"content":  base64.StdEncoding.EncodeToString(fileBytes),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return ParseResumeResponse{}, fmt.Errorf("marshal resume parse payload: %w", err)
	}

	parseCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(parseCtx, p.pythonPath, "-c", resumeParsePythonScript)
	cmd.Stdin = bytes.NewReader(payloadJSON)
	cmd.Env = append(
		os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return ParseResumeResponse{}, fmt.Errorf("parse resume file: %s", message)
	}

	var result ParseResumeResponse
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return ParseResumeResponse{}, fmt.Errorf("decode parsed resume text: %w", err)
	}

	result.FileName = name
	result.Text = strings.TrimSpace(result.Text)
	if result.Text == "" {
		return ParseResumeResponse{}, errors.New("no readable text found in the resume")
	}

	return result, nil
}

const resumeParsePythonScript = `
import base64
import io
import json
import sys
from pathlib import Path

from pypdf import PdfReader
from docx import Document

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8")


def read_payload():
    raw = sys.stdin.buffer.read()
    if not raw:
        raise ValueError("empty stdin payload")
    return json.loads(raw.decode("utf-8"))


def parse_pdf(data: bytes) -> str:
    reader = PdfReader(io.BytesIO(data))
    parts = []
    for page in reader.pages:
        text = page.extract_text() or ""
        text = text.strip()
        if text:
            parts.append(text)
    return "\n\n".join(parts)


def parse_docx(data: bytes) -> str:
    document = Document(io.BytesIO(data))
    parts = []
    for para in document.paragraphs:
        text = para.text.strip()
        if text:
            parts.append(text)
    for table in document.tables:
        for row in table.rows:
            cells = []
            for cell in row.cells:
                cell_text = "\n".join(
                    paragraph.text.strip()
                    for paragraph in cell.paragraphs
                    if paragraph.text.strip()
                ).strip()
                if cell_text:
                    cells.append(cell_text)
            if cells:
                parts.append(" | ".join(cells))
    return "\n".join(parts)


def main():
    payload = read_payload()
    file_name = payload.get("fileName", "")
    ext = Path(file_name).suffix.lower()
    data = base64.b64decode(payload.get("content", ""))

    if ext == ".pdf":
        text = parse_pdf(data)
    elif ext == ".docx":
        text = parse_docx(data)
    else:
        text = data.decode("utf-8", errors="ignore")

    result = {
        "fileName": file_name,
        "text": text.strip(),
    }
    sys.stdout.write(json.dumps(result, ensure_ascii=True))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        sys.stderr.write(str(exc))
        sys.exit(1)
`

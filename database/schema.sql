CREATE DATABASE IF NOT EXISTS ai_interview
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE ai_interview;

CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NULL,
    nickname VARCHAR(60) NOT NULL,
    avatar_url MEDIUMTEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(120) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(60) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE candidate_profiles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    real_name VARCHAR(60) NULL,
    target_role VARCHAR(80) NOT NULL,
    experience_level VARCHAR(30) NOT NULL,
    years_of_experience DECIMAL(4,1) NOT NULL DEFAULT 0,
    education_level VARCHAR(40) NULL,
    self_intro TEXT NULL,
    strengths TEXT NULL,
    weak_points TEXT NULL,
    metadata JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_candidate_profiles_user_id FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE resumes (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    candidate_profile_id BIGINT NULL,
    name VARCHAR(120) NOT NULL,
    source_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    file_url VARCHAR(255) NULL,
    raw_text LONGTEXT NULL,
    parsed_content JSON NULL,
    version_no INT NOT NULL DEFAULT 1,
    is_default TINYINT(1) NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_resumes_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_resumes_candidate_profile_id FOREIGN KEY (candidate_profile_id) REFERENCES candidate_profiles(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE job_targets (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    candidate_profile_id BIGINT NULL,
    resume_id BIGINT NULL,
    company_name VARCHAR(120) NULL,
    job_title VARCHAR(120) NOT NULL,
    job_category VARCHAR(60) NOT NULL,
    level_code VARCHAR(30) NOT NULL,
    interview_type VARCHAR(30) NOT NULL,
    jd_text LONGTEXT NULL,
    jd_keywords JSON NULL,
    custom_requirements TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_job_targets_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_job_targets_candidate_profile_id FOREIGN KEY (candidate_profile_id) REFERENCES candidate_profiles(id),
    CONSTRAINT fk_job_targets_resume_id FOREIGN KEY (resume_id) REFERENCES resumes(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE interview_plans (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    job_title VARCHAR(120) NOT NULL,
    job_category VARCHAR(60) NOT NULL,
    level_code VARCHAR(30) NOT NULL,
    interview_type VARCHAR(30) NOT NULL,
    interview_mode VARCHAR(30) NOT NULL DEFAULT 'text',
    difficulty_level VARCHAR(20) NOT NULL DEFAULT 'medium',
    question_count INT NOT NULL DEFAULT 5,
    interviewer_style VARCHAR(30) NOT NULL DEFAULT 'balanced',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_interview_plans_user_id FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE interview_sessions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    candidate_profile_id BIGINT NULL,
    job_target_id BIGINT NULL,
    resume_id BIGINT NULL,
    job_title VARCHAR(120) NOT NULL,
    session_name VARCHAR(120) NOT NULL,
    round_type VARCHAR(30) NOT NULL,
    interview_mode VARCHAR(30) NOT NULL DEFAULT 'text',
    interviewer_style VARCHAR(30) NOT NULL DEFAULT 'balanced',
    difficulty_level VARCHAR(20) NOT NULL DEFAULT 'medium',
    question_count INT NOT NULL DEFAULT 5,
    current_question_no INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    started_at DATETIME NULL,
    ended_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_interview_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_interview_sessions_candidate_profile_id FOREIGN KEY (candidate_profile_id) REFERENCES candidate_profiles(id),
    CONSTRAINT fk_interview_sessions_job_target_id FOREIGN KEY (job_target_id) REFERENCES job_targets(id),
    CONSTRAINT fk_interview_sessions_resume_id FOREIGN KEY (resume_id) REFERENCES resumes(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE interview_questions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    question_no INT NOT NULL,
    question_type VARCHAR(30) NOT NULL,
    assessment_dimension VARCHAR(40) NOT NULL,
    prompt TEXT NOT NULL,
    expected_points JSON NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'ai',
    is_follow_up TINYINT(1) NOT NULL DEFAULT 0,
    parent_question_id BIGINT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_interview_questions_session_no UNIQUE (session_id, question_no),
    CONSTRAINT fk_interview_questions_session_id FOREIGN KEY (session_id) REFERENCES interview_sessions(id),
    CONSTRAINT fk_interview_questions_parent_question_id FOREIGN KEY (parent_question_id) REFERENCES interview_questions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE interview_answers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    answer_text LONGTEXT NULL,
    answer_audio_url VARCHAR(255) NULL,
    answer_duration_seconds INT NULL,
    submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_interview_answers_session_id FOREIGN KEY (session_id) REFERENCES interview_sessions(id),
    CONSTRAINT fk_interview_answers_question_id FOREIGN KEY (question_id) REFERENCES interview_questions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE answer_feedbacks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    answer_id BIGINT NOT NULL,
    overall_score DECIMAL(5,2) NOT NULL,
    accuracy_score DECIMAL(5,2) NULL,
    clarity_score DECIMAL(5,2) NULL,
    depth_score DECIMAL(5,2) NULL,
    communication_score DECIMAL(5,2) NULL,
    strengths TEXT NULL,
    issues TEXT NULL,
    improvement_suggestion TEXT NULL,
    follow_up_question TEXT NULL,
    feedback_payload JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_answer_feedbacks_answer_id UNIQUE (answer_id),
    CONSTRAINT fk_answer_feedbacks_answer_id FOREIGN KEY (answer_id) REFERENCES interview_answers(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE interview_reports (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    session_id BIGINT NOT NULL,
    overall_score DECIMAL(5,2) NOT NULL,
    performance_summary TEXT NOT NULL,
    strengths_summary TEXT NULL,
    weakness_summary TEXT NULL,
    recommended_actions TEXT NULL,
    next_plan JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_interview_reports_session_id UNIQUE (session_id),
    CONSTRAINT fk_interview_reports_session_id FOREIGN KEY (session_id) REFERENCES interview_sessions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE user_ai_provider_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    provider_code VARCHAR(30) NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    api_key_masked VARCHAR(64) NOT NULL,
    model VARCHAR(80) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    is_enabled TINYINT(1) NOT NULL DEFAULT 1,
    last_test_ok TINYINT(1) NOT NULL DEFAULT 0,
    last_tested_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT uq_user_ai_provider_configs_user_provider UNIQUE (user_id, provider_code),
    CONSTRAINT fk_user_ai_provider_configs_user_id FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE report_dimension_scores (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    report_id BIGINT NOT NULL,
    dimension_code VARCHAR(40) NOT NULL,
    score DECIMAL(5,2) NOT NULL,
    comment TEXT NULL,
    CONSTRAINT uq_report_dimension_scores UNIQUE (report_id, dimension_code),
    CONSTRAINT fk_report_dimension_scores_report_id FOREIGN KEY (report_id) REFERENCES interview_reports(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_candidate_profiles_user_id ON candidate_profiles(user_id);
CREATE INDEX idx_resumes_user_id ON resumes(user_id);
CREATE INDEX idx_job_targets_user_id ON job_targets(user_id);
CREATE INDEX idx_interview_plans_user_id ON interview_plans(user_id);
CREATE INDEX idx_interview_sessions_user_id ON interview_sessions(user_id);
CREATE INDEX idx_interview_sessions_job_target_id ON interview_sessions(job_target_id);
CREATE INDEX idx_interview_sessions_status ON interview_sessions(status);
CREATE INDEX idx_interview_questions_session_id ON interview_questions(session_id);
CREATE INDEX idx_interview_answers_session_id ON interview_answers(session_id);
CREATE INDEX idx_interview_answers_question_id ON interview_answers(question_id);
CREATE INDEX idx_user_ai_provider_configs_user_id ON user_ai_provider_configs(user_id);
CREATE INDEX idx_report_dimension_scores_report_id ON report_dimension_scores(report_id);

-- AI Symptom Checker - Initial Database Schema
-- Run this in the Supabase SQL Editor to manually create tables

-- Enable UUID extension if not enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users Table (Patients)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    age INTEGER,
    gender VARCHAR(10),
    known_allergies TEXT,
    pre_existing_conditions TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 2. Doctors Table
CREATE TYPE doctor_status AS ENUM ('pending', 'active', 'suspended');
CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    specialization TEXT,
    credentials TEXT,
    invite_token TEXT,
    invite_expiry TIMESTAMP WITH TIME ZONE,
    status doctor_status DEFAULT 'pending',
    is_active BOOLEAN DEFAULT TRUE,
    invited_by_admin_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_doctors_email ON doctors(email);
CREATE INDEX IF NOT EXISTS idx_doctors_invite_token ON doctors(invite_token);
CREATE INDEX IF NOT EXISTS idx_doctors_deleted_at ON doctors(deleted_at);

-- 3. Admins Table
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_admins_email ON admins(email);
CREATE INDEX IF NOT EXISTS idx_admins_username ON admins(username);
CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins(deleted_at);

-- 4. Symptom Sessions Table
CREATE TYPE urgency_level AS ENUM ('emergency', 'see_doctor', 'self_care');
CREATE TABLE IF NOT EXISTS symptom_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    raw_input TEXT NOT NULL,
    processed_input TEXT,
    urgency_level urgency_level,
    ai_provider TEXT,
    is_flagged_for_review BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_symptom_sessions_user_id ON symptom_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_symptom_sessions_deleted_at ON symptom_sessions(deleted_at);

-- 5. Diagnoses Table (linked to Session)
CREATE TABLE IF NOT EXISTS diagnoses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES symptom_sessions(id) ON DELETE CASCADE,
    condition_name TEXT NOT NULL,
    description TEXT,
    confidence TEXT,
    common_causes TEXT,
    health_advice TEXT,
    rank INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_diagnoses_session_id ON diagnoses(session_id);

-- 6. Consultations Table
CREATE TYPE consultation_status AS ENUM ('pending', 'answered', 'closed');
CREATE TYPE consultation_urgency AS ENUM ('routine', 'soon', 'urgent');
CREATE TABLE IF NOT EXISTS consultations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE SET NULL,
    session_id UUID REFERENCES symptom_sessions(id) ON DELETE SET NULL,
    symptoms TEXT NOT NULL,
    patient_note TEXT,
    urgency consultation_urgency DEFAULT 'routine',
    status consultation_status DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_consultations_user_id ON consultations(user_id);
CREATE INDEX IF NOT EXISTS idx_consultations_doctor_id ON consultations(doctor_id);
CREATE INDEX IF NOT EXISTS idx_consultations_deleted_at ON consultations(deleted_at);

-- 7. Consultation Replies Table
CREATE TABLE IF NOT EXISTS consultation_replies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    consultation_id UUID NOT NULL REFERENCES consultations(id) ON DELETE CASCADE,
    doctor_id UUID NOT NULL REFERENCES doctors(id) ON DELETE CASCADE,
    reply_text TEXT NOT NULL,
    recommendation TEXT,
    is_ai_correction BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_consultation_replies_consultation_id ON consultation_replies(consultation_id);

-- 8. Notifications Table
CREATE TYPE notification_type AS ENUM ('consultation_answered', 'case_reviewed', 'doctor_approved', 'system_alert');
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    doctor_id UUID REFERENCES doctors(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    ref_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_doctor_id ON notifications(doctor_id);

-- 9. Feedback Table
CREATE TABLE IF NOT EXISTS feedbacks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES symptom_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    helpful BOOLEAN,
    note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_feedbacks_session_id ON feedbacks(session_id);

-- 10. Knowledge Base Table
CREATE TYPE knowledge_status AS ENUM ('active', 'pending', 'rejected');
CREATE TABLE IF NOT EXISTS knowledge_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    disease_name TEXT NOT NULL,
    symptoms TEXT NOT NULL,
    description TEXT,
    causes TEXT,
    advice TEXT,
    source TEXT,
    is_core_dataset BOOLEAN DEFAULT FALSE,
    status knowledge_status DEFAULT 'active',
    submitted_by UUID REFERENCES doctors(id) ON DELETE SET NULL,
    approved_by UUID REFERENCES admins(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_knowledge_disease ON knowledge_entries(disease_name);
CREATE INDEX IF NOT EXISTS idx_knowledge_deleted_at ON knowledge_entries(deleted_at);

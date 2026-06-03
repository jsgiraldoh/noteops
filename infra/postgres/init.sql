-- Extensiones
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Usuarios del sistema (docentes / admin)
CREATE TABLE IF NOT EXISTS users (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name   TEXT NOT NULL,
  email       TEXT UNIQUE NOT NULL,
  password    TEXT NOT NULL,
  role        TEXT NOT NULL DEFAULT 'teacher', -- teacher | admin
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Materias
CREATE TABLE IF NOT EXISTS subjects (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL,
  period      TEXT NOT NULL,
  group_name  TEXT,
  faculty     TEXT,
  teacher_id  UUID REFERENCES users(id),
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Cortes (configurable por materia: 1, 2, 3... N)
CREATE TABLE IF NOT EXISTS cuts (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id  UUID REFERENCES subjects(id) ON DELETE CASCADE,
  number      INT NOT NULL,
  name        TEXT NOT NULL DEFAULT '',
  weight      DECIMAL(5,4) NOT NULL,
  UNIQUE(subject_id, number),
  CHECK(weight > 0 AND weight <= 1)
);

-- Actividades dentro de cada corte (N1, N2, Parcial...)
CREATE TABLE IF NOT EXISTS activities (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cut_id      UUID REFERENCES cuts(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  weight      DECIMAL(5,4) NOT NULL,
  scheduled_at DATE,
  CHECK(weight > 0 AND weight <= 1)
);

-- Estudiantes
CREATE TABLE IF NOT EXISTS students (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name   TEXT NOT NULL,
  email       TEXT UNIQUE NOT NULL,
  code        TEXT UNIQUE,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Inscripción estudiante ↔ materia
CREATE TABLE IF NOT EXISTS enrollments (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id  UUID REFERENCES students(id) ON DELETE CASCADE,
  subject_id  UUID REFERENCES subjects(id) ON DELETE CASCADE,
  UNIQUE(student_id, subject_id)
);

-- Notas por actividad
CREATE TABLE IF NOT EXISTS grades (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  enrollment_id UUID REFERENCES enrollments(id) ON DELETE CASCADE,
  activity_id   UUID REFERENCES activities(id) ON DELETE CASCADE,
  value         DECIMAL(3,1) CHECK(value >= 0 AND value <= 5),
  comment       TEXT,
  recorded_at   TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(enrollment_id, activity_id)
);

-- Sesiones de clase (para el reloj + reserva de espacios)
CREATE TABLE IF NOT EXISTS sessions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  subject_id   UUID REFERENCES subjects(id) ON DELETE CASCADE,
  starts_at    TIMESTAMPTZ NOT NULL,
  duration_min INT NOT NULL DEFAULT 120,
  slot_min     INT NOT NULL DEFAULT 120, -- duración de cada espacio (5/10/20/120)
  room         TEXT,
  active       BOOLEAN DEFAULT false,
  created_at   TIMESTAMPTZ DEFAULT NOW()
);

-- Espacios dentro de una sesión
CREATE TABLE IF NOT EXISTS slots (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id  UUID REFERENCES sessions(id) ON DELETE CASCADE,
  number      INT NOT NULL,
  starts_at   TIMESTAMPTZ NOT NULL,
  duration_min INT NOT NULL,
  student_id  UUID REFERENCES students(id),
  reserved_at TIMESTAMPTZ,
  UNIQUE(session_id, number)
);

-- Vista: nota definitiva calculada automáticamente
CREATE OR REPLACE VIEW student_final_grades AS
SELECT
  e.id          AS enrollment_id,
  e.student_id,
  e.subject_id,
  ROUND(
    COALESCE(SUM(g.value * a.weight * c.weight), 0)::NUMERIC, 2
  )             AS final_grade
FROM enrollments e
LEFT JOIN grades g     ON g.enrollment_id = e.id
LEFT JOIN activities a ON a.id = g.activity_id
LEFT JOIN cuts c       ON c.id = a.cut_id
GROUP BY e.id, e.student_id, e.subject_id;

-- Índices útiles
CREATE INDEX IF NOT EXISTS idx_grades_enrollment ON grades(enrollment_id);
CREATE INDEX IF NOT EXISTS idx_activities_cut    ON activities(cut_id);
CREATE INDEX IF NOT EXISTS idx_slots_session     ON slots(session_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_subj  ON enrollments(subject_id);

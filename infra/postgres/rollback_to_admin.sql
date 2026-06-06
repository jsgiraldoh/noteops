-- ─────────────────────────────────────────────────────────────────────────────
-- rollback_to_admin.sql
-- Deja la base de datos limpia: elimina todos los datos académicos y conserva
-- únicamente el usuario administrador (admin@noteops.local).
--
-- Uso:
--   make rollback
--   o directamente:
--   docker compose exec postgres psql -U noteops -d noteops -f /rollback_to_admin.sql
-- ─────────────────────────────────────────────────────────────────────────────

BEGIN;

-- 1. Notas (dependen de enrollments y activities)
DELETE FROM grades;

-- 2. Slots (dependen de sessions)
DELETE FROM slots;

-- 3. Sessions (dependen de subjects)
DELETE FROM sessions;

-- 4. Enrollments (dependen de students y subjects)
DELETE FROM enrollments;

-- 5. Activities (dependen de cuts)
DELETE FROM activities;

-- 6. Cuts (dependen de subjects)
DELETE FROM cuts;

-- 7. Subjects
DELETE FROM subjects;

-- 8. Students
DELETE FROM students;

-- 9. Usuarios — conservar solo el administrador
DELETE FROM users WHERE email != 'admin@noteops.local';

COMMIT;

-- Verificación
SELECT 'grades'      AS tabla, COUNT(*) AS filas FROM grades
UNION ALL SELECT 'slots',       COUNT(*) FROM slots
UNION ALL SELECT 'sessions',    COUNT(*) FROM sessions
UNION ALL SELECT 'enrollments', COUNT(*) FROM enrollments
UNION ALL SELECT 'activities',  COUNT(*) FROM activities
UNION ALL SELECT 'cuts',        COUNT(*) FROM cuts
UNION ALL SELECT 'subjects',    COUNT(*) FROM subjects
UNION ALL SELECT 'students',    COUNT(*) FROM students
UNION ALL SELECT 'users',       COUNT(*) FROM users;

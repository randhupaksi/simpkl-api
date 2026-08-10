DELETE rp FROM role_permissions rp
JOIN permissions p ON p.id = rp.permission_id
WHERE p.module = 'automation';
DELETE FROM permissions WHERE module = 'automation';
DROP TABLE IF EXISTS generated_documents;
DROP TABLE IF EXISTS document_generation_batches;
DROP TABLE IF EXISTS letter_sequences;
DROP TABLE IF EXISTS document_templates;
DROP TABLE IF EXISTS signatories;
DROP TABLE IF EXISTS school_profiles;

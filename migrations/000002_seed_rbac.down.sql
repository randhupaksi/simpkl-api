DELETE FROM document_types WHERE code IN ('acceptance_letter','parent_permission','introduction_letter','mou','supervisor_assignment');
DELETE FROM role_permissions;
DELETE FROM roles WHERE is_system = TRUE;
DELETE FROM permissions;

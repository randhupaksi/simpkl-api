# Document Automation

SIMPKL turns verified placement data into official PKL administration files.
The engine is deterministic and does not use generative AI for official values.

## Outputs

- editable Office Open XML (`.docx`) letters;
- A4 portrait PDF letters;
- styled placement recap workbooks (`.xlsx`);
- one ZIP package for every individual or batch generation request.

The seeded templates cover introduction/application letters, placement
statements, supervisor assignments, parent/guardian consent, and placement
recaps. Administrators can create institution-specific templates and new
versions through the API or web application.

## Supported template placeholders

| Placeholder | Source |
| --- | --- |
| `{{academic_year}}`, `{{period_name}}` | PKL period |
| `{{student_name}}`, `{{student_nis}}`, `{{student_nisn}}` | student |
| `{{class_name}}`, `{{major_name}}` | academic master data |
| `{{parent_name}}`, `{{parent_phone}}` | student guardian data |
| `{{company_name}}`, `{{company_address}}`, `{{company_city}}` | company/placement |
| `{{company_contact_name}}`, `{{company_contact_position}}` | company contact |
| `{{supervisor_name}}`, `{{supervisor_employee_number}}`, `{{supervisor_position}}` | supervisor |
| `{{placement_division}}`, `{{placement_position}}` | placement |
| `{{placement_start}}`, `{{placement_end}}`, `{{letter_date}}` | formatted dates |

Number patterns support `{{sequence}}`, `{{code}}`, `{{month}}`,
`{{month_roman}}`, and `{{year}}`.

## API workflow

```text
filters or placement IDs
  -> completeness preview
  -> active template + transactional letter sequence
  -> immutable source snapshot
  -> DOCX/PDF/XLSX render
  -> private storage
  -> generated document metadata + SHA-256
  -> batch ZIP + audit event
```

Main routes are under `/api/v1/document-automation`: profile, signatories,
templates, preview, generate, batches, generated documents, and authenticated
downloads.

## Operational safeguards

- Preview validates source completeness before generation.
- Number increments are protected by a row lock inside a transaction.
- Optional foreign keys are written as SQL `NULL`, never as empty UUIDs.
- Generated files remain private and require download permission.
- Every file stores template version, source snapshot, SHA-256 checksum, user,
  timestamp, number, format, MIME type, and size.
- Updating a template creates a new version and keeps prior versions for audit.

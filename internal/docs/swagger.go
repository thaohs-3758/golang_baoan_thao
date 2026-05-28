package docs

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Service Management System API",
    "description": "API documentation for the Service Management System backend.\n\n**Authentication:** Most citizen endpoints require a Bearer JWT access token obtained from POST /api/auth/login. The refresh token is stored as an HttpOnly cookie and renewed via POST /api/auth/refresh.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local development"
    }
  ],
  "tags": [
    { "name": "Auth",            "description": "Register, login, token refresh, logout" },
    { "name": "Citizen Profile", "description": "Citizen's own profile (view & update)" },
    { "name": "Applications",    "description": "Submit and track public-service applications" },
    { "name": "Service Catalog", "description": "Browse available service types" },
    { "name": "Notifications",   "description": "Citizen notifications" },
    { "name": "Admin — Citizens",      "description": "Admin: list, export and import citizen accounts" },
    { "name": "Admin — Departments",   "description": "Admin: export and import departments" },
    { "name": "Admin — Staff",         "description": "Admin: export and import staff accounts" },
    { "name": "Admin — Service Types", "description": "Admin: export and import service types" },
    { "name": "Admin — Applications",  "description": "Admin: export applications" }
  ],
  "paths": {
    "/api/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Register a new citizen account",
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/RegisterRequest" },
              "example": {
                "name": "Nguyen Van A",
                "email": "citizen@example.com",
                "password": "secret123",
                "citizen_id_number": "012345678901"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Account created — returns the new user object",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UserResponse" } } }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "409": {
            "description": "Email already registered",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          }
        }
      }
    },
    "/api/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Login",
        "description": "Returns a short-lived **access token** in the JSON body and sets the refresh token in an HttpOnly cookie (name: refresh_token, 7 days).",
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/LoginRequest" },
              "example": { "email": "citizen@example.com", "password": "secret123" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login successful",
            "headers": {
              "Set-Cookie": {
                "description": "HttpOnly refresh token cookie (7 days)",
                "schema": { "type": "string", "example": "refresh_token=eyJ...; Path=/; HttpOnly; SameSite=Lax" }
              }
            },
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/LoginResponse" } } }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": {
            "description": "Wrong email or password",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          },
          "403": {
            "description": "Account is blocked",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          }
        }
      }
    },
    "/api/auth/refresh": {
      "post": {
        "tags": ["Auth"],
        "summary": "Refresh access token",
        "description": "Reads the refresh_token HttpOnly cookie and returns a new 15-minute access token. No request body needed.",
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "responses": {
          "200": {
            "description": "New access token issued",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/TokenResponse" } } }
          },
          "401": {
            "description": "Cookie missing, token invalid, or token expired",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          }
        }
      }
    },
    "/api/auth/logout": {
      "post": {
        "tags": ["Auth"],
        "summary": "Logout",
        "description": "Clears the refresh_token HttpOnly cookie. The client should also discard its access token.",
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "responses": {
          "200": {
            "description": "Logged out successfully",
            "headers": {
              "Set-Cookie": {
                "description": "Expires the refresh token cookie",
                "schema": { "type": "string", "example": "refresh_token=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax" }
              }
            },
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } }
          }
        }
      }
    },
    "/api/citizens/me": {
      "get": {
        "tags": ["Citizen Profile"],
        "summary": "Get my profile",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "responses": {
          "200": {
            "description": "Profile retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProfileResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" }
        }
      },
      "put": {
        "tags": ["Citizen Profile"],
        "summary": "Update my profile",
        "description": "All fields are optional. citizen_id_number is immutable and cannot be changed via this endpoint.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/UpdateProfileRequest" },
              "example": {
                "name": "Nguyen Van B",
                "phone": "0901234567",
                "gender": "male",
                "email_notification_enabled": true
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Profile updated",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProfileResponse" } } }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" }
        }
      }
    },
    "/api/citizens/me/password": {
      "put": {
        "tags": ["Citizen Profile"],
        "summary": "Change my password",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": { "$ref": "#/components/schemas/ChangeMyPasswordRequest" }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Password changed successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" },
          "422": {
            "description": "Confirmation mismatch, new password equals current password, or validation failed",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          }
        }
      }
    },
    "/api/citizens/me/applications": {
      "get": {
        "tags": ["Applications"],
        "summary": "List my applications",
        "description": "Returns a paginated list of the authenticated citizen's applications, ordered by submitted_at descending.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "page",  "in": "query", "schema": { "type": "integer", "minimum": 1, "default": 1 }, "description": "Page number" },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "minimum": 1, "maximum": 100, "default": 10 }, "description": "Items per page" }
        ],
        "responses": {
          "200": {
            "description": "Applications retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationListResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" }
        }
      },
      "post": {
        "tags": ["Applications"],
        "summary": "Submit a new application",
        "description": "Uses multipart/form-data. The 'data' field must be a JSON string matching SubmitApplicationRequest. Attachments are optional (PDF / JPG / PNG only, max 10 files, 10 MiB per file, 30 MiB total).",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["data"],
                "properties": {
                  "data": {
                    "type": "string",
                    "description": "JSON string — see SubmitApplicationRequest schema",
                    "example": "{\"service_type_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"submitted_data\":{\"full_name\":\"Nguyen Van A\",\"dob\":\"1990-01-01\"}}"
                  },
                  "attachments[]": {
                    "type": "array",
                    "items": { "type": "string", "format": "binary" },
                    "description": "Optional attachments (PDF / JPG / PNG)"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Application submitted — confirmation email sent to citizen",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationDetailResponse" } } }
          },
          "400": { "$ref": "#/components/responses/BadRequest" },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "422": {
            "description": "Service not found, inactive, missing required form field, or attachment limit exceeded",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
          }
        }
      }
    },
    "/api/citizens/me/applications/{id}": {
      "get": {
        "tags": ["Applications"],
        "summary": "Get application detail",
        "description": "Returns the full detail of a single application. Only the owning citizen can access it.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" }, "description": "Application ID" }
        ],
        "responses": {
          "200": {
            "description": "Application detail retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationDetailResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" }
        }
      }
    },
    "/api/citizens/me/applications/{id}/status-history": {
      "get": {
        "tags": ["Applications"],
        "summary": "Get my application status history",
        "description": "Returns status timeline for a citizen-owned application. Supports polling with since (RFC3339).",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } },
          { "name": "page",  "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 10, "maximum": 100 } },
          { "name": "since", "in": "query", "schema": { "type": "string", "format": "date-time" }, "description": "Only return logs after this timestamp (RFC3339)" }
        ],
        "responses": {
          "200": {
            "description": "Status history retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationStatusHistoryResponse" } } }
          },
          "400": { "description": "Invalid query parameter", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Application not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/me/applications/{id}/supplements": {
      "post": {
        "tags": ["Applications"],
        "summary": "Upload supplement attachments",
        "description": "Upload additional documents for a citizen-owned application. Allowed only when status is processing or need_more_info.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["attachments[]"],
                "properties": {
                  "attachments[]": {
                    "type": "array",
                    "items": { "type": "string", "format": "binary" },
                    "description": "Supplement file attachments (PDF/JPG/PNG, max 10 files, 10 MiB each, 30 MiB total)"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Supplement attachments uploaded successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationAttachmentsResponse" } } }
          },
          "400": { "description": "Invalid request or empty attachments", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Application not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "422": { "description": "Upload not allowed by status or file constraints", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/services": {
      "get": {
        "tags": ["Service Catalog"],
        "summary": "List available service types",
        "description": "Returns only active service types. Supports search by name and filter by category.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "page",     "in": "query", "schema": { "type": "integer", "minimum": 1, "default": 1 } },
          { "name": "limit",    "in": "query", "schema": { "type": "integer", "minimum": 1, "maximum": 100, "default": 20 } },
          { "name": "category", "in": "query", "schema": { "type": "string" }, "description": "Filter by category name (exact match)" },
          { "name": "search",   "in": "query", "schema": { "type": "string" }, "description": "Search by service name (case-insensitive)" }
        ],
        "responses": {
          "200": {
            "description": "Service list retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ServiceListResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" }
        }
      }
    },
    "/admin/citizens": {
      "get": {
        "tags": ["Admin — Citizens"],
        "summary": "List all citizen accounts",
        "description": "Returns a paginated list of citizens with their application count. Admin / manager / staff access only.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "page",  "in": "query", "schema": { "type": "integer", "minimum": 1, "default": 1 } },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "minimum": 1, "maximum": 500, "default": 20 } }
        ],
        "responses": {
          "200": { "description": "Citizen list page (HTML)", "content": { "text/html": { "schema": { "type": "string" } } } },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/users/export/citizens": {
      "get": {
        "tags": ["Admin — Citizens"],
        "summary": "Export all citizens as CSV",
        "description": "Streams all citizen rows (so_cccd, ho_ten, email, so_dien_thoai, dia_chi, ngay_sinh, tong_ho_so) as a UTF-8 BOM CSV file. Filename: citizens.csv. Super Admin only.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV file download",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"citizens.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/citizens/template": {
      "get": {
        "tags": ["Admin — Citizens"],
        "summary": "Download CSV import template for citizens",
        "description": "Returns a CSV file with the expected header row: so_cccd,ho_ten,email. Super Admin only.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV template file",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"citizens_template.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/citizens/import": {
      "post": {
        "tags": ["Admin — Citizens"],
        "summary": "Import citizens from CSV",
        "description": "Accepts a multipart/form-data upload with field 'file' containing a CSV. Expected columns: so_cccd, ho_ten, email. All rows are imported atomically — if any row fails validation the entire import is rolled back and errors are returned. Initial password is set to the citizen's CCCD number.",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["file"],
                "properties": {
                  "file": { "type": "string", "format": "binary", "description": "CSV file with header row: so_cccd,ho_ten,email" }
                }
              }
            }
          }
        },
        "responses": {
          "303": { "description": "Import successful — redirects back to /admin/citizens" },
          "400": { "description": "No file uploaded or CSV is empty", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" },
          "422": { "description": "One or more rows failed validation — full list of per-row errors returned, no data saved", "content": { "text/html": { "schema": { "type": "string" } } } }
        }
      }
    },
    "/admin/departments/export": {
      "get": {
        "tags": ["Admin — Departments"],
        "summary": "Export all departments as CSV",
        "description": "Streams all departments (ten, mo_ta, ma_code) as a UTF-8 BOM CSV file. Filename: departments.csv.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV file download",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"departments.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/departments/template": {
      "get": {
        "tags": ["Admin — Departments"],
        "summary": "Download CSV import template for departments",
        "description": "Returns a CSV file with the expected header row: ten,mo_ta,ma_code. Manager + Super Admin.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV template file",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"departments_template.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/departments/import": {
      "post": {
        "tags": ["Admin — Departments"],
        "summary": "Import departments from CSV",
        "description": "Accepts a multipart/form-data upload. Expected CSV columns: ten, mo_ta, ma_code. All rows imported atomically.",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["file"],
                "properties": {
                  "file": { "type": "string", "format": "binary", "description": "CSV file with header row: ten,mo_ta,ma_code" }
                }
              }
            }
          }
        },
        "responses": {
          "303": { "description": "Import successful — redirects back to /admin/departments" },
          "400": { "description": "No file uploaded or CSV is empty", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" },
          "422": { "description": "Validation errors — no data saved", "content": { "text/html": { "schema": { "type": "string" } } } }
        }
      }
    },
    "/admin/users/export/staff": {
      "get": {
        "tags": ["Admin — Staff"],
        "summary": "Export all staff accounts as CSV",
        "description": "Streams all staff users (ho_ten, email, so_dien_thoai, dia_chi, vai_tro, trang_thai) as a UTF-8 BOM CSV file. Filename: users.csv. Super Admin only.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV file download",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"users.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/users/template": {
      "get": {
        "tags": ["Admin — Staff"],
        "summary": "Download CSV import template for staff accounts",
        "description": "Returns a CSV file with the expected header row: ho_ten,email,so_cccd,so_dien_thoai,vai_tro,ma_phong_ban. Super Admin only.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV template file",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"users_template.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/users/import": {
      "post": {
        "tags": ["Admin — Staff"],
        "summary": "Import staff accounts from CSV",
        "description": "Accepts a multipart/form-data upload. Expected CSV columns: ho_ten, email, so_cccd, so_dien_thoai, vai_tro, ma_phong_ban. Valid roles: staff, manager, super_admin. Default password: Aa@123456.",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["file"],
                "properties": {
                  "file": { "type": "string", "format": "binary", "description": "CSV file with header row: ho_ten,email,so_cccd,so_dien_thoai,vai_tro,ma_phong_ban" }
                }
              }
            }
          }
        },
        "responses": {
          "303": { "description": "Import successful — redirects back to /admin/users" },
          "400": { "description": "No file uploaded or CSV is empty", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" },
          "422": { "description": "Validation errors — no data saved", "content": { "text/html": { "schema": { "type": "string" } } } }
        }
      }
    },
    "/admin/service-types/export": {
      "get": {
        "tags": ["Admin — Service Types"],
        "summary": "Export all service types as CSV",
        "description": "Streams all service types (ten, mo_ta, thoi_gian_xu_ly_ngay, phi, ma_phong_ban) as a UTF-8 BOM CSV file. Filename: service_types.csv.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV file download",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"service_types.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/service-types/template": {
      "get": {
        "tags": ["Admin — Service Types"],
        "summary": "Download CSV import template for service types",
        "description": "Returns a CSV file with the expected header row: ten,mo_ta,thoi_gian_xu_ly_ngay,phi,ma_phong_ban. Super Admin only.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV template file",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"service_types_template.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/admin/service-types/import": {
      "post": {
        "tags": ["Admin — Service Types"],
        "summary": "Import service types from CSV",
        "description": "Accepts a multipart/form-data upload. Expected CSV columns: ten, mo_ta, thoi_gian_xu_ly_ngay, phi, ma_phong_ban. Service code is auto-generated from the name.",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["file"],
                "properties": {
                  "file": { "type": "string", "format": "binary", "description": "CSV file with header row: ten,mo_ta,thoi_gian_xu_ly_ngay,phi,ma_phong_ban" }
                }
              }
            }
          }
        },
        "responses": {
          "303": { "description": "Import successful — redirects back to /admin/service-types" },
          "400": { "description": "No file uploaded or CSV is empty", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" },
          "422": { "description": "Validation errors — no data saved", "content": { "text/html": { "schema": { "type": "string" } } } }
        }
      }
    },
    "/admin/applications/export": {
      "get": {
        "tags": ["Admin — Applications"],
        "summary": "Export all applications as CSV",
        "description": "Streams all applications (ma_ho_so, ten_cong_dan, loai_dich_vu, phong_ban, trang_thai, ngay_nop, ngay_hoan_thanh, can_bo_xu_ly) as a UTF-8 BOM CSV file. Filename: applications.csv.",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": {
            "description": "CSV file download",
            "headers": {
              "Content-Disposition": { "schema": { "type": "string", "example": "attachment; filename=\"applications.csv\"" } }
            },
            "content": { "text/csv": { "schema": { "type": "string", "format": "binary" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "403": { "$ref": "#/components/responses/Forbidden" }
        }
      }
    },
    "/api/citizens/services/{id}": {
      "get": {
        "tags": ["Service Catalog"],
        "summary": "Get service type detail",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" }, "description": "Service type ID" }
        ],
        "responses": {
          "200": {
            "description": "Service type retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ServiceDetailResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "404": { "$ref": "#/components/responses/NotFound" }
        }
      }
    },
    "/api/citizens/me/notifications": {
      "get": {
        "tags": ["Notifications"],
        "summary": "List my notifications",
        "description": "Returns a paginated list of the authenticated citizen's notifications, ordered by created_at descending.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "page",    "in": "query", "schema": { "type": "integer", "minimum": 1, "default": 1 }, "description": "Page number" },
          { "name": "limit",   "in": "query", "schema": { "type": "integer", "minimum": 1, "maximum": 100, "default": 10 }, "description": "Items per page" },
          { "name": "type",    "in": "query", "schema": { "type": "string" }, "description": "Filter by notification type" },
          { "name": "is_read", "in": "query", "schema": { "type": "boolean" }, "description": "Filter by read status (true / false)" }
        ],
        "responses": {
          "200": {
            "description": "Notifications retrieved",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/NotificationListResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "description": "Internal server error", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/me/notifications/read-all": {
      "put": {
        "tags": ["Notifications"],
        "summary": "Mark all notifications as read",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "responses": {
          "200": {
            "description": "All notifications marked as read",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "description": "Internal server error", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/me/notifications/{id}/read": {
      "put": {
        "tags": ["Notifications"],
        "summary": "Mark a single notification as read",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" }, "description": "Notification ID" }
        ],
        "responses": {
          "200": {
            "description": "Notification marked as read",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/MessageResponse" } } }
          },
          "401": { "$ref": "#/components/responses/Unauthorized" },
          "500": { "description": "Internal server error", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    }
  },
  "components": {
    "responses": {
      "BadRequest": {
        "description": "Validation error or malformed request body",
        "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
      },
      "Unauthorized": {
        "description": "Missing or invalid Bearer token",
        "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
      },
      "Forbidden": {
        "description": "Authenticated but not authorized (wrong role)",
        "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
      },
      "NotFound": {
        "description": "Resource not found",
        "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } }
      }
    },
    "schemas": {
      "RegisterRequest": {
        "type": "object",
        "required": ["name", "email", "password", "citizen_id_number"],
        "properties": {
          "name": {
            "type": "string",
            "example": "Nguyen Van A"
          },
          "email": {
            "type": "string",
            "format": "email",
            "example": "user@example.com"
          },
          "password": {
            "type": "string",
            "minLength": 6,
            "example": "123456"
          },
          "citizen_id_number": {
            "type": "string",
            "description": "12-digit numeric citizen ID number",
            "example": "123456789012"
          }
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email":    { "type": "string", "format": "email", "example": "citizen@example.com" },
          "password": { "type": "string", "example": "secret123" }
        }
      },
      "SubmitApplicationRequest": {
        "type": "object",
        "required": ["service_type_id", "submitted_data"],
        "properties": {
          "service_type_id": { "type": "string", "format": "uuid", "description": "ID of the service type to apply for" },
          "submitted_data":  { "type": "object", "additionalProperties": true, "description": "Free-form JSON object matching the service type's form schema" }
        }
      },
      "UpdateProfileRequest": {
        "type": "object",
        "description": "All fields are optional. Omitted fields are not changed.",
        "properties": {
          "name":                       { "type": "string", "minLength": 1 },
          "phone":                      { "type": "string", "example": "0901234567" },
          "address":                    { "type": "string" },
          "gender":                     { "type": "string", "example": "male" },
          "permanent_address":          { "type": "string" },
          "date_of_birth":              { "type": "string", "format": "date-time", "nullable": true },
          "email_notification_enabled": { "type": "boolean" }
        }
      },
      "ChangeMyPasswordRequest": {
        "type": "object",
        "required": ["current_password", "new_password", "confirm_new_password"],
        "properties": {
          "current_password":     { "type": "string", "minLength": 6, "maxLength": 72, "example": "oldpass123" },
          "new_password":         { "type": "string", "minLength": 6, "maxLength": 72, "example": "newpass123" },
          "confirm_new_password": { "type": "string", "minLength": 6, "maxLength": 72, "example": "newpass123" }
        }
      },
      "User": {
        "type": "object",
        "properties": {
          "id":           { "type": "string", "format": "uuid" },
          "name":         { "type": "string" },
          "email":        { "type": "string", "format": "email" },
          "phone":        { "type": "string" },
          "address":      { "type": "string" },
          "role":         { "type": "string", "enum": ["citizen", "staff", "manager", "super_admin"] },
          "status":       { "type": "string", "enum": ["active", "blocked"] },
          "created_at":   { "type": "string", "format": "date-time" },
          "updated_at":   { "type": "string", "format": "date-time" }
        }
      },
      "UserResponse": {
        "type": "object",
        "properties": {
          "user": {
            "$ref": "#/components/schemas/User"
          }
        }
      },
      "LoginResponse": {
        "type": "object",
        "properties": {
          "user":  { "$ref": "#/components/schemas/User" },
          "token": { "type": "string", "description": "JWT access token (15 min)", "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
        }
      },
      "TokenResponse": {
        "type": "object",
        "properties": {
          "token": { "type": "string", "description": "New JWT access token (15 min)" }
        }
      },
      "MessageResponse": {
        "type": "object",
        "properties": {
          "message": { "type": "string" }
        }
      },
      "CitizenProfile": {
        "type": "object",
        "properties": {
          "user_id":                    { "type": "string", "format": "uuid" },
          "name":                       { "type": "string" },
          "email":                      { "type": "string", "format": "email" },
          "phone":                      { "type": "string" },
          "address":                    { "type": "string" },
          "citizen_id_number":          { "type": "string", "example": "012345678901" },
          "gender":                     { "type": "string" },
          "permanent_address":          { "type": "string" },
          "date_of_birth":              { "type": "string", "format": "date-time", "nullable": true },
          "email_notification_enabled": { "type": "boolean" },
          "created_at":                 { "type": "string", "format": "date-time" },
          "updated_at":                 { "type": "string", "format": "date-time" }
        }
      },
      "ProfileResponse": {
        "type": "object",
        "properties": {
          "profile": { "$ref": "#/components/schemas/CitizenProfile" }
        }
      },
      "ApplicationItem": {
        "description": "Summary row returned in the list endpoint",
        "type": "object",
        "properties": {
          "id":                { "type": "string", "format": "uuid" },
          "application_code":  { "type": "string", "example": "APP-20260521-A1B2C3" },
          "service_type_id":   { "type": "string", "format": "uuid" },
          "service_type_name": { "type": "string" },
          "status":            { "type": "string", "enum": ["received", "processing", "need_more_info", "approved", "rejected"] },
          "submitted_at":      { "type": "string", "format": "date-time" }
        }
      },
      "ApplicationAttachment": {
        "type": "object",
        "properties": {
          "id":        { "type": "string", "format": "uuid" },
          "file_name": { "type": "string" },
          "file_url":  { "type": "string", "description": "Public URL to download the file" },
          "file_type": { "type": "string", "example": "application/pdf" },
          "file_size": { "type": "integer", "nullable": true, "description": "File size in bytes" }
        }
      },
      "ApplicationDetail": {
        "description": "Full detail of a single application including attachments and submitted form data",
        "type": "object",
        "properties": {
          "id":                { "type": "string", "format": "uuid" },
          "application_code":  { "type": "string", "example": "APP-20260521-A1B2C3" },
          "service_type_id":   { "type": "string", "format": "uuid" },
          "service_type_name": { "type": "string" },
          "status":            { "type": "string", "enum": ["received", "processing", "need_more_info", "approved", "rejected"] },
          "submitted_data":    { "type": "object", "additionalProperties": true, "description": "Free-form data submitted with the application" },
          "submitted_at":      { "type": "string", "format": "date-time" },
          "attachments":       { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationAttachment" } }
        }
      },
      "ApplicationListResponse": {
        "type": "object",
        "properties": {
          "applications": { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationItem" } },
          "pagination":   { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ApplicationDetailResponse": {
        "type": "object",
        "properties": {
          "application": { "$ref": "#/components/schemas/ApplicationDetail" }
        }
      },
      "ApplicationStatusHistoryItem": {
        "type": "object",
        "properties": {
          "id":                   { "type": "string", "format": "uuid" },
          "old_status":           { "type": "string", "enum": ["received", "processing", "need_more_info", "approved", "rejected"], "nullable": true },
          "new_status":           { "type": "string", "enum": ["received", "processing", "need_more_info", "approved", "rejected"] },
          "note":                 { "type": "string" },
          "changed_by_user_id":   { "type": "string", "format": "uuid", "nullable": true },
          "changed_by_user_name": { "type": "string", "nullable": true },
          "created_at":           { "type": "string", "format": "date-time" }
        }
      },
      "ApplicationStatusHistoryResponse": {
        "type": "object",
        "properties": {
          "status_history": { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationStatusHistoryItem" } },
          "pagination":     { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ApplicationAttachmentsResponse": {
        "type": "object",
        "properties": {
          "attachments": { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationAttachment" } }
        }
      },
      "ServiceType": {
        "type": "object",
        "properties": {
          "id":          { "type": "string", "format": "uuid" },
          "name":        { "type": "string" },
          "code":        { "type": "string" },
          "description": { "type": "string" },
          "category":    { "type": "string" },
          "is_active":   { "type": "boolean" },
          "created_at":  { "type": "string", "format": "date-time" }
        }
      },
      "ServiceListResponse": {
        "type": "object",
        "properties": {
          "services":   { "type": "array", "items": { "$ref": "#/components/schemas/ServiceType" } },
          "pagination": { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ServiceDetailResponse": {
        "type": "object",
        "properties": {
          "service": { "$ref": "#/components/schemas/ServiceType" }
        }
      },
      "Pagination": {
        "type": "object",
        "properties": {
          "page":  { "type": "integer", "example": 1 },
          "limit": { "type": "integer", "example": 10 },
          "total": { "type": "integer", "example": 42 }
        }
      },
      "NotificationResponse": {
        "type": "object",
        "properties": {
          "id":             { "type": "string", "format": "uuid" },
          "application_id": { "type": "string", "format": "uuid", "nullable": true },
          "title":          { "type": "string" },
          "message":        { "type": "string" },
          "type":           { "type": "string", "description": "Notification type, e.g. application_received, status_updated, supplement_requested" },
          "is_read":        { "type": "boolean" },
          "read_at":        { "type": "string", "format": "date-time", "nullable": true },
          "created_at":     { "type": "string", "format": "date-time" }
        }
      },
      "NotificationListResponse": {
        "type": "object",
        "properties": {
          "notifications": { "type": "array", "items": { "$ref": "#/components/schemas/NotificationResponse" } },
          "pagination":    { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ErrorDetail": {
        "type": "object",
        "properties": {
          "field":   { "type": "string", "description": "Field name (only present for validation errors)", "example": "email" },
          "code":    { "type": "string", "description": "i18n error key", "example": "validation.email" },
          "message": { "type": "string", "description": "Human-readable message in the requested language", "example": "Email không hợp lệ" }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "code":   { "type": "integer", "description": "HTTP status code", "example": 400 },
          "errors": { "type": "array", "items": { "$ref": "#/components/schemas/ErrorDetail" } }
        }
      }
    },
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Access token obtained from /api/auth/login (expires in 15 min). Use /api/auth/refresh to renew."
      }
    }
  }
}`

const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Service Management System API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

func SetupSwaggerRoutes(e *echo.Echo) {
	e.GET("/swagger", swaggerIndex)
	e.GET("/swagger/", swaggerIndex)
	e.GET("/swagger/doc.json", swaggerSpec)
}

func swaggerIndex(c *echo.Context) error {
	return c.HTML(http.StatusOK, swaggerUI)
}

func swaggerSpec(c *echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", []byte(openAPISpec))
}

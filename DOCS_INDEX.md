# 📚 Lambra Documentation Index

Guide untuk semua dokumentasi proyek Lambra.

---

## 🚀 Quick Start Documentation

**Mulai di sini jika Anda baru dengan Lambra:**

1. **[README.md](README.md)** - Overview proyek dan quick start
   - Pengenalan fitur-fitur Lambra
   - Tech stack dan arsitektur
   - Quick start guide
   - Cara membuat relations (visual)

2. **[SETUP.md](SETUP.md)** - Panduan setup lengkap
   - Prerequisites (Docker, Docker Compose)
   - Step-by-step installation
   - Environment configuration
   - Troubleshooting

3. **[dist/README.md](dist/README.md)** - Panduan untuk end-user
   - Quick start untuk penggunaan
   - Database configuration
   - Port configuration

---

## 🎯 Feature Documentation

**Dokumentasi fitur-fitur spesifik:**

### Visual Relations Feature 🎉
**[FEATURE_VISUAL_RELATIONS.md](FEATURE_VISUAL_RELATIONS.md)**
- Fitur terbaru: Drag-and-drop relation creation
- 6 fase implementasi (semua complete ✅)
- Progress tracking dan deliverables
- Testing checklist

### Relations Technical Specification
**[RELATIONS_TABLE_SPEC.md](RELATIONS_TABLE_SPEC.md)**
- Technical spec untuk relations table
- Naming conventions & semantic meaning
- Contoh implementasi untuk setiap tipe relasi
- FAQ dan best practices

---

## 🧪 Testing & QA

### Testing Checklist
**[TESTING_CHECKLIST.md](TESTING_CHECKLIST.md)**
- 9 fase testing scenarios
- 21+ test cases
- Critical path testing
- Bug report template

---

## 🤖 Development Documentation

### AI Assistant Guide
**[CLAUDE.md](CLAUDE.md)**
- Guide untuk Claude Code (AI assistant)
- Project overview dan commands
- Backend architecture patterns
- Frontend structure
- API routes
- Current development status

### Progress Tracker
**[PROGRESS.md](PROGRESS.md)**
- Overall project roadmap
- Phase-by-phase progress
- Detailed completion status
- Next steps

---

## 📖 Documentation Structure

```
lambra/
├── README.md                      # Main project README (Indonesian)
├── SETUP.md                       # Setup guide (Indonesian)
├── DOCS_INDEX.md                  # This file - Documentation index
├── CLAUDE.md                      # AI assistant guide
├── PROGRESS.md                    # Project progress tracker
├── FEATURE_VISUAL_RELATIONS.md    # Visual relations feature doc
├── RELATIONS_TABLE_SPEC.md        # Relations technical spec
├── TESTING_CHECKLIST.md           # QA testing checklist
└── dist/
    └── README.md                  # Distribution readme (English)
```

---

## 🎓 Reading Paths

### Path 1: New Developer
1. README.md - Understand the project
2. SETUP.md - Set up development environment
3. CLAUDE.md - Development patterns and commands
4. PROGRESS.md - Current status and roadmap

### Path 2: Product Manager
1. README.md - Feature overview
2. FEATURE_VISUAL_RELATIONS.md - Latest feature
3. TESTING_CHECKLIST.md - Testing status
4. PROGRESS.md - Overall roadmap

### Path 3: QA Tester
1. SETUP.md - Environment setup
2. TESTING_CHECKLIST.md - Testing scenarios
3. FEATURE_VISUAL_RELATIONS.md - Feature details
4. RELATIONS_TABLE_SPEC.md - Technical details

### Path 4: End User
1. dist/README.md - Quick start
2. README.md - Features overview
3. SETUP.md - Setup (if self-hosting)

---

## 📝 Key Concepts

### Visual Relations Feature
- **Drag-and-Drop**: Create relations by dragging between entities in diagram
- **4 Relation Types**: belongsTo (pink), hasOne (purple), hasMany (blue), manyToMany (orange)
- **Interactive**: Hover to see details, click to edit/delete
- **Smart Generation**: Auto-generates FK constraints and junction tables
- **Full CRUD**: Complete create, read, update, delete operations

### Dual Identifier Strategy
- **Internal ID**: BIGINT for FK relationships (not exposed in API)
- **External UUID**: CHAR(36) exposed as "id" in API
- Derived from UUID v7

### Soft Deletes
- All deletions set `deleted_at` timestamp
- No hard deletes (except cascade)

---

## 🔗 Quick Links

- **GitHub**: (Repository URL)
- **Issues**: (Issues URL)
- **Documentation**: https://lambra.dev/docs (if available)

---

## 📅 Maintenance Notes

**Last Updated:** 2026-03-04

**Recent Updates:**
- ✅ Fixed backend utility functions (toSnakeCase, pluralize)
- ✅ Fixed frontend API response handling in DatabaseDiagram
- ✅ Updated README with Visual Relations feature
- ✅ Created documentation index (this file)

**Todo:**
- [ ] Add screenshots/videos for Visual Relations feature
- [ ] Create video tutorial for setup
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Create contribution guidelines

---

*For questions or updates, contact the development team.*

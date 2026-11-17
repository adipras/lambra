# Konten PDF Submit Hasil — Lambra

> Format submit wajib sesuai website: PDF di Google Drive (akses terbuka), paling lambat **15 November 2026**.
> Angka bertanda **[UKUR]** wajib diganti hasil pengukuran tim sendiri sebelum submit.

---

## 01 — Problem

Setiap permintaan sistem baru menuntut tim membangun microservice dari nol: model, repository, service, handler, DTO, migrasi database, sampai Dockerfile. Pekerjaan ini:

- **Repetitif** — 70–80% adalah boilerplate CRUD yang strukturnya selalu sama.
- **Lambat** — kurang lebih 4–8 jam per service sebelum bisa dites.
- **Tidak konsisten** — tiap developer punya gaya penulisan sendiri; hasil copy-paste codebase lama menyulitkan review dan maintenance.

## 02 — Current Process

1. Permintaan sistem diterima → System Analyst menganalisa kebutuhan sistem → UI/UX mendesain sistem yang sudah didefinisikan System Analyst → developer (back-end) menyiapkan scaffolding manual ( struktur folder, dependensi, konfigurasi DB).
2. Menulis ulang layer yang sama: model → repository → service → handler → DTO → migrasi SQL.
3. Menulis Dockerfile & docker-compose, uji koneksi DB, debug konfigurasi.
4. Review kode yang sebagian besar boilerplate — menyita waktu reviewer.
5. Deploy manual, tidak ada snapshot/rollback baku; audit dilakukan lewat ingatan/chat.

Total waktu ke service pertama berjalan: [UKUR] jam–hari per orang. Konsistensi bergantung disiplin individu.

## 03 — Solution

**Lambra** — platform generator microservice internal berbasis web:

1. **Definisi visual**: entity + field + relasi (belongsTo/hasOne/hasMany/manyToMany) dibuat lewat diagram drag-and-drop; FK dan constraint ON DELETE/UPDATE digenerate otomatis.
2. **Generate kode lengkap**: satu klik menghasilkan project Golang utuh — model, repository, service, handler, DTO dengan partial update, migrasi SQL dengan foreign key, OpenAPI/Postman export, Dockerfile multi-stage, docker-compose.
3. **Deploy satu klik**: build + jalankan service sebagai container terisolasi dengan database dedicated; progres per-step terlihat real-time (SSE).
4. **Keamanan bawaan**: JWT + bcrypt untuk platform; service generated mengekspos UUID (internal ID tidak pernah bocor); database terisolasi per service.
5. **Tata kelola**: setiap deploy membuat snapshot; audit trail (siapa-kapan-apa) pada project, deployment, snapshot; rollback penuh teruji (entity dipulihkan + redeploy otomatis, data bertahan).

## 04 — Impact (Before vs After)

| Aspek | Before (manual) | After (Lambra) |
|---|---|---|
| Waktu ke service CRUD berjalan | [UKUR] 4–8 jam | **< 5 menit** (terukur: definisi ±1 menit + build/deploy 40–60 detik) |
| Konsistensi kode | Bergantung individu | **100% dari template tereview** — perbaikan template berlaku ke semua service berikutnya |
| Rollback | Manual, tidak baku | Snapshot 1-klik + redeploy otomatis |
| Audit | Tidak ada | created_by + timestamp + log per-step di semua operasi |
| Onboarding developer internal ke service baru | Baca codebase manual | Definisi entity terbaca dari diagram |

**Perhitungan dampak tahunan** [UKUR sesuai volume tim]: bila tim membuat N=20 service internal/tahun × hemat ±X jam = ±20X jam developer yang dialihkan ke pekerjaan bernilai lebih tinggi.

## 05 — Governance

- **Audit trail**: tabel project/deployment/snapshot mencatat pelaku (JWT-authenticated user) dan waktu; log deployment per-step tersimpan dan dapat di-stream.
- **Snapshot & rollback**: setiap deploy = snapshot; rollback memulihkan definisi entity dan me-redeploy service (teruji end-to-end, data existing bertahan).
- **Konsistensi = compliance**: semua service generated mengikuti standar sama (dual-ID strategy, soft delete, format response seragam) — memudahkan audit keamanan dan review.
- **Proses tim**: pengembangan platform sendiri melalui branch + pull request review (jejak review dapat ditunjukkan) — praktik governance diterapkan pada diri sendiri.

## 06 — Security (Bagaimana mencegah solusi menjadi security risk?)

- **Autentikasi**: seluruh API platform dilindungi JWT (HS256, expiry 24 jam); password disimpan sebagai hash bcrypt; endpoint publik hanya health-check.
- **Isolasi**: tiap service generated berjalan di container + network + database sendiri; kebocoran satu service tidak menyebar ke lainnya.
- **Anti data leakage**: API eksternal hanya menerima/mengembalikan UUID — internal BIGINT ID tidak pernah diekspos; password DB tidak pernah dikembalikan response API.
- **Generated code hardening**: validasi input (binding + validator tags) di setiap handler generated; foreign key constraint menjaga integritas data di level database.
- **Roadmap**: role-based access control (admin/developer/viewer), integrasi GitLab CI dengan secret management.

---

## Checklist sebelum submit
- [ ] Semua angka [UKUR] diganti hasil pengukuran nyata tim
- [ ] Screenshot: diagram relation, deploy progress, deployment history dengan created_by, service healthy + curl response
- [ ] Link backup video demo (opsional, tambahkan di Impact)
- [ ] Export ke PDF, upload Google Drive, set akses "Anyone with the link"
- [ ] Submit sebelum 15 November 2026

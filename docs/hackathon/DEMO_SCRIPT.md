# Lambra — Demo Script (Hackathon BSI UII 2026)

> Durasi inti: **5 menit** + Q&A. Versi pendek 3 menit ada di bawah.
> Track: **Faster — Automasi & Efisiensi** (dengan bukti Governed & Safer).
> Tim: Bacalah — presentasi oleh 1 orang, 1 orang driver demo, 1 orang siap jawab teknis.

---

## Persiapan (H-1, ±15 menit)

### 1. Reset environment ke kondisi bersih

```bash
make down -v          # bersihkan semua container + volume
make up               # stack baru
make migrate-up       # jalankan 2x jika error "Duplicate foreign key" (runner belum state-aware);
                      # users table bisa juga di-apply manual:
                      # docker exec -i lambra-mysql mysql -ulambra -plambra_secret lambra_db \
                      #   < backend/migrations/005_create_users_table.up.sql
```

### 2. Verifikasi pre-flight (semua harus hijau sebelum demo)

```bash
curl -s http://localhost:8080/health                 # {"status":"ok"} semacamnya
curl -s http://localhost:5173 -o /dev/null -w "%{http_code}\n"   # 200
# login admin/admin123 via UI browser — PASTIKAN frontend branch sudah di-merge
```

### 3. Siapkan data demo
- Browser: tab 1 = Lambra UI (sudah login), tab 2 = terminal dengan curl siap
- Terminal: sudah login dan `TOKEN` tersimpan (script di bawah)
- (Opsional) 1 service "Payroll" sudah ter-deploy sebagai bukti sebelum mulai

### 4. Rekam backup video
```bash
# macOS: Cmd+Shift+5, record fullscreen, simpan sebagai demo-backup.mov
```
**Wajib** — kalau live demo gagal, mainkan video.

---

## Script Inti (5 menit)

### [0:00–0:40] Problem + hook (bicara, tanpa klik)
> "Setiap ada kebutuhan sistem internal baru — form pengajuan, pencatatan aset, laporan — tim kita menulis ulang boilerplate yang sama: model, repository, service, handler, DTO, migrasi database, Dockerfile. Butuh **jam sampai hari** per service, kualitas tidak konsisten karena tiap orang punya gaya sendiri, dan tidak ada jejak siapa mengubah apa."

> "Lambra: definisi entity lewat diagram, klik deploy, service Go lengkap jadi dalam **menit** — dengan audit trail."

### [0:40–1:30] Login + buat service (UI)
1. Logout dulu (tunjukkan halaman login) → login `admin`
2. **Create New Service**: nama `Asset`, namespace `asset-ns`, DB config default (host `lambra-mysql`, user `lambra`, password `lambra_secret`, db `asset_db`)
3. Klik Create → tunjukkan validasi koneksi DB otomatis

> "Akses dikontrol JWT — semua operasi tercatat siapa pelakunya. Bukan 'system' anonim."

### [1:30–2:40] Definisi entity + relation (UI)
1. Buat entity `Laptop`: fields `name` (string, required), `serial` (string, required, unique), `notes` (text, optional)
2. Buat entity `Employee`: fields `name` (string, required)
3. Diagram view → drag relation: `Laptop` **belongsTo** `Employee` (CASCADE)

> "Relasi antar entity dibuat visual — FK, constraint ON DELETE, semuanya digenerate otomatis, termasuk migrasi SQL dan foreign key-nya."

### [2:40–3:40] Deploy + live progress (UI)
1. Klik **Deploy** → biarkan progress modal terlihat (generate → docker build → start)
2. Selesai → tunjukkan URL service + status healthy

> "Dari diagram ke service berjalan. Waktu build tadi sekitar 40–60 detik — tanpa menulis satu baris kode."

### [3:40–4:20] Buktikan service hidup (terminal)
```bash
PORT=<port-dari-response-deploy>
curl -s http://localhost:$PORT/health
curl -s -X POST http://localhost:$PORT/laptops -H "Content-Type: application/json" \
  -d '{"name":"ThinkPad T14","serial":"TP-001","employee_uuid":null}'
curl -s http://localhost:$PORT/laptops
```

> "CRUD lengkap, validasi, UUID sebagai identitas publik — internal ID tidak pernah bocor."

### [4:20–5:00] Governance close (UI + DB)
1. Deployment History → tunjukkan log per-step + **created_by: admin**
2. Snapshot → Create snapshot → (boleh singkat: jelaskan rollback sudah teruji)
3. Tutup:

> "Setiap deploy tersimpan sebagai snapshot dengan pelaku dan waktu — bisa di-rollback penuh, termasuk redeploy otomatis. Dokumentasi, audit trail, dan konsistensi kode bawaan platform. Dari jam menjadi menit — faster, safer, governed."

---

## Versi 3 menit (kalau waktu dipangkas)
Login → create service → 1 entity → deploy → curl POST → deployment history → closing.

## Bank jawaban Q&A (kemungkinan pertanyaan juri)

| Pertanyaan | Jawaban inti |
|---|---|
| "Bagaimana mencegah jadi security risk?" | JWT semua endpoint, bcrypt, per-service DB terisolasi, UUID eksternal (internal ID tidak diekspos), soft-delete + audit siapa-melakukan-apa. Roadmap: role-based access. |
| "Kalau service generated punya celah?" | Kode digenerate dari template yang sudah direview sekali — perbaikan di satu template berlaku ke semua service berikutnya, beda dengan copy-paste manual. |
| "Kenapa tidak pakai AI?" | Fokus track Faster dengan bukti terukur; AI-assisted schema design ada di roadmap. |
| "Bagaimana kalau deploy gagal?" | Status tercatat failed + log per-step; snapshot rollback teruji end-to-end (entity + data bertahan). |
| "Skala ke production?" | Pattern sama, ganti target deploy docker-compose → Kubernetes (namespace field sudah disiapkan). GitLab CI di roadmap. |

## Risiko demo & mitigasi
1. **Docker build lambat/saat demo** → pre-pull image, jalankan sekali pagi hari; fallback video.
2. **Port berubah tiap deploy** → selalu baca port dari response deploy, jangan hafal.
3. **Entity UUID berubah setelah rollback** → UI harus refetch (sudah dikomunikasikan ke tim frontend).
4. **Lupa password DB MySQL** → `lambra_secret` (tertulis di form default).

# GitLab Migration Guide (แยก Repository Backend และ Frontend)

คู่มือสำหรับการตั้งค่า Git แยกโปรเจกต์ API (Backend) และ Frontend ขึ้น GitLab แยกเป็น 2 Repositories ต่างกัน

---

## ขั้นตอนที่ 0: ลบ Git เดิมที่คลุมนอกสุด
รันคำสั่งนี้ที่โฟลเดอร์หลักนอกสุด (`Omnilogs`) เพื่อลบโฟลเดอร์ `.git` เดิมออก (รันผ่าน PowerShell)

```powershell
Remove-Item -Recurse -Force .git
```

---

## ขั้นตอนที่ 1: จัดการฝั่ง API (Backend)

1. เข้าไปยังโฟลเดอร์ backend:
   ```bash
   cd backend
   ```

2. เริ่มต้นระบบ Git เฉพาะโฟลเดอร์นี้:
   ```bash
   git init
   ```

3. แอดโค้ดทั้งหมดและบันทึก Commit:
   ```bash
   git add .
   git commit -m "Initial commit for backend API"
   ```

4. ตั้งชื่อ Branch หลัก และสลับไปที่ Branch `development`:
   ```bash
   git branch -M main
   git checkout -b development
   ```

5. เชื่อมต่อไปยัง GitLab API repository:
   ```bash
   git remote add origin git@gitlab.dev.nextate.tech:activity-logs/omni-logs.git
   ```

6. ดึงข้อมูลเริ่มต้นและทำการ Push โค้ดขึ้นไป:
   *(หมายเหตุ: บน GitLab ต้องกดเพิ่ม README หรือไฟล์ตั้งต้นเพื่อให้เกิด Default branch ก่อนถึงจะทำได้)*
   ```bash
   git push -u origin development
   ```

---

## ขั้นตอนที่ 2: จัดการฝั่ง Frontend

1. สลับมาที่โฟลเดอร์ frontend:
   ```bash
   cd ../frontend
   ```

2. เริ่มต้นระบบ Git เฉพาะโฟลเดอร์นี้:
   ```bash
   git init
   ```

3. แอดโค้ดทั้งหมดและบันทึก Commit:
   ```bash
   git add .
   git commit -m "Initial commit for frontend"
   ```

4. ตั้งชื่อ Branch หลัก:
   ```bash
   git branch -M main
   ```

5. เชื่อมต่อไปยัง GitLab Frontend repository (ใช้ SSH URL):
   ```bash
   git remote add origin git@gitlab.dev.nextate.tech:activity-logs/omni-logs-frontend.git
   ```

6. ซิงค์โปรเจกต์และเชื่อมโยงประวัติเข้าด้วยกัน:
   ```bash
   git pull origin main --allow-unrelated-histories --no-edit
   ```
   *(แก้ไขความขัดแย้งในไฟล์ `README.md` หากมี)*
   ```bash
   git add README.md
   git commit -m "Merge branch 'main' of gitlab"
   ```

7. สร้าง Branch `development` และอัปโหลดโค้ดขึ้น GitLab:
   ```bash
   git checkout -b development
   git push -u origin development
   ```

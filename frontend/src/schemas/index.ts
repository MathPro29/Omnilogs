import { z } from 'zod/v4';

// ===== Login Schema =====
export const loginSchema = z.object({
  username: z
    .string()
    .min(1, 'กรุณากรอกชื่อผู้ใช้')
    .min(3, 'ชื่อผู้ใช้ต้องมีอย่างน้อย 3 ตัวอักษร'),
  password: z
    .string()
    .min(1, 'กรุณากรอกรหัสผ่าน')
    .min(6, 'รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร'),
});

export type LoginFormData = z.infer<typeof loginSchema>;

// ===== Create User Schema =====
export const createUserSchema = z.object({
  username: z
    .string()
    .min(1, 'กรุณากรอกชื่อผู้ใช้')
    .min(3, 'ชื่อผู้ใช้ต้องมีอย่างน้อย 3 ตัวอักษร')
    .max(50, 'ชื่อผู้ใช้ต้องไม่เกิน 50 ตัวอักษร'),
  email: z
    .email('รูปแบบอีเมลไม่ถูกต้อง'),
  fullName: z
    .string()
    .min(1, 'กรุณากรอกชื่อ-นามสกุล'),
  phone: z
    .string()
    .regex(/^[0-9]{9,10}$/, 'เบอร์โทรศัพท์ไม่ถูกต้อง')
    .optional()
    .or(z.literal('')),
  department: z.string().optional(),
  position: z.string().optional(),
  role: z.string().min(1, 'กรุณาเลือกบทบาท'),
  password: z
    .string()
    .min(1, 'กรุณากรอกรหัสผ่าน')
    .min(8, 'รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร'),
});

export type CreateUserFormData = z.infer<typeof createUserSchema>;

// ===== Edit User Schema =====
export const editUserSchema = z.object({
  email: z
    .email('รูปแบบอีเมลไม่ถูกต้อง'),
  fullName: z
    .string()
    .min(1, 'กรุณากรอกชื่อ-นามสกุล'),
  phone: z
    .string()
    .regex(/^[0-9]{9,10}$/, 'เบอร์โทรศัพท์ไม่ถูกต้อง')
    .optional()
    .or(z.literal('')),
  department: z.string().optional(),
  position: z.string().optional(),
  role: z.string().min(1, 'กรุณาเลือกบทบาท'),
  status: z.enum(['active', 'inactive', 'pending']),
});

export type EditUserFormData = z.infer<typeof editUserSchema>;

// ===== Settings Schema =====
export const settingsSchema = z.object({
  siteName: z.string().min(1, 'กรุณากรอกชื่อระบบ'),
  siteDescription: z.string().optional(),
  maintenanceMode: z.boolean(),
  allowRegistration: z.boolean(),
  defaultRole: z.string().min(1, 'กรุณาเลือกบทบาทเริ่มต้น'),
  sessionTimeout: z
    .number()
    .min(5, 'ต้องไม่น้อยกว่า 5 นาที')
    .max(1440, 'ต้องไม่เกิน 1440 นาที'),
  maxLoginAttempts: z
    .number()
    .min(1, 'ต้องไม่น้อยกว่า 1 ครั้ง')
    .max(20, 'ต้องไม่เกิน 20 ครั้ง'),
});

export type SettingsFormData = z.infer<typeof settingsSchema>;

// ===== User Filter Schema =====
export const userFilterSchema = z.object({
  search: z.string().optional(),
  status: z.enum(['active', 'inactive', 'pending', '']).optional(),
  department: z.string().optional(),
  role: z.string().optional(),
});

export type UserFilterFormData = z.infer<typeof userFilterSchema>;

// ===== Register Schema =====
export const registerSchema = z
  .object({
    username: z
      .string()
      .min(3, 'ชื่อผู้ใช้ต้องมีอย่างน้อย 3 ตัวอักษร')
      .max(50, 'ชื่อผู้ใช้ต้องไม่เกิน 50 ตัวอักษร')
      .optional()
      .or(z.literal('')),
    first_name: z
      .string()
      .min(1, 'กรุณากรอกชื่อจริง'),
    last_name: z
      .string()
      .min(1, 'กรุณากรอกนามสกุล'),
    email: z
      .string()
      .min(1, 'กรุณากรอกอีเมล')
      .email('รูปแบบอีเมลไม่ถูกต้อง'),
    phone_number: z
      .string()
      .regex(/^[0-9]{8,10}$/, 'เบอร์โทรศัพท์ต้องเป็นตัวเลข 8-10 หลัก')
      .optional()
      .or(z.literal('')),
    password: z
      .string()
      .min(1, 'กรุณากรอกรหัสผ่าน')
      .min(8, 'รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร'),
    confirm_password: z
      .string()
      .min(1, 'กรุณากรอกยืนยันรหัสผ่าน'),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: 'รหัสผ่านและการยืนยันรหัสผ่านไม่ตรงกัน',
    path: ['confirm_password'],
  });

export type RegisterFormData = z.infer<typeof registerSchema>;

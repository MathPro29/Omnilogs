import { z } from "zod";

export const productInformationSchema = z.object({
  product_name: z.string().min(1, "Product Name is required").max(100, "Product Name must be under 100 characters"),
  description: z.string().optional().default(""),
  owner_id: z.number().optional(),
  is_active: z.boolean().default(true),
});

export type ProductInformationFormValues = z.infer<typeof productInformationSchema>;

import { z } from "zod";

export const featureSchema = z.object({
  category_name: z.string().min(1, "Feature Name is required"),
  description: z.string().optional().default(""),
  is_active: z.boolean().default(true),
});

export type FeatureFormValues = z.infer<typeof featureSchema>;

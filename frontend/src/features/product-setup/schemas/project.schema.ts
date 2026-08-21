import { z } from "zod";

export const projectSchema = z.object({
  project_name: z.string().min(1, "Project Name is required"),
  description: z.string().optional().default(""),
  is_active: z.boolean().default(true),
});

export type ProjectFormValues = z.infer<typeof projectSchema>;

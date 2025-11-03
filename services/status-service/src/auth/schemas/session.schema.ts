import { z } from 'zod';

export const SessionDataSchema = z.object({
  user_id: z.union([z.string(), z.number()]).transform((val) => {
    const num = typeof val === 'string' ? parseInt(val, 10) : val;
    if (isNaN(num) || num <= 0) {
      throw new Error('Invalid user_id');
    }
    return num;
  }),
  username: z.string().min(1).max(16),
});

export type SessionData = z.infer<typeof SessionDataSchema>;
export type AuthUser = {
  userId: number;
  username: string;
};

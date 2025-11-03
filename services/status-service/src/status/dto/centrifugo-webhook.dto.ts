import { z } from 'zod';

export const CentrifugoWebhookDto = z.object({
  event: z.enum(['connect', 'disconnect']),
  user: z.string(),
});

export type CentrifugoWebhook = z.infer<typeof CentrifugoWebhookDto>;

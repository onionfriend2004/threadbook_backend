import { SessionDataSchema, AuthUser } from './schemas/session.schema';
import { RedisService } from '../shared/redis/redis.service';
import { Injectable } from '@nestjs/common';
import { ZodError } from 'zod';

function safeJsonParse(input: string): unknown {
  return JSON.parse(input);
}

@Injectable()
export class AuthService {
  constructor(private redisService: RedisService) {}

  async authenticate(sid: string): Promise<AuthUser | null> {
    const sessionData = await this.redisService.get(`session_id:${sid}`);
    if (!sessionData) {
      return null;
    }

    let parsed: unknown;
    try {
      parsed = safeJsonParse(sessionData);
    } catch {
      return null;
    }

    try {
      const validated = SessionDataSchema.parse(parsed);
      return {
        userId: validated.user_id,
        username: validated.username,
      };
    } catch (error) {
      if (error instanceof ZodError) {
        console.warn('Invalid session data:', error);
      }
      return null;
    }
  }
}

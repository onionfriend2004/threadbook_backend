import { Injectable, Logger } from '@nestjs/common';
import { UpdateStatusDto } from './dto/update-status.dto';
import { PrismaService } from '../shared/prisma/prisma.service';
import { OnlineStatus } from '@prisma/client';
import { RedisService } from 'src/shared/redis/redis.service';
import { UserStatus } from './schemas/user-status.schema';

@Injectable()
export class StatusService {
  private readonly logger = new Logger(StatusService.name);

  constructor(
    private readonly prisma: PrismaService,
    private readonly redis: RedisService,
  ) {}

  // === Online status (Redis) ===

  async markOnline(userId: number): Promise<void> {
    try {
      const pipeline = this.redis.multi();
      pipeline.set(`online-status:user:${userId}`, '1', { EX: 70 });
      pipeline.set(`last-seen:user:${userId}`, Date.now().toString(), {
        EX: 30 * 24 * 3600 * 12, // +- год
      });
      await pipeline.exec();
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to mark user ${userId} as online`, err);
    }
  }

  async markOffline(userId: number): Promise<void> {
    try {
      const pipeline = this.redis.multi();
      pipeline.del(`online-status:user:${userId}`);
      pipeline.set(`last-seen:user:${userId}`, Date.now().toString(), {
        EX: 30 * 24 * 3600,
      });
      await pipeline.exec();
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to mark user ${userId} as offline`, err);
    }
  }

  async isOnline(userId: number): Promise<boolean> {
    try {
      return await this.redis.exists(`online-status:user:${userId}`);
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(
        `Failed to check online status for user ${userId}`,
        err,
      );
      return false;
    }
  }

  // === Custom status (PostgreSQL) ===

  async updateCustomStatus(
    userId: number,
    username: string,
    dto: UpdateStatusDto,
  ): Promise<void> {
    try {
      await this.prisma.onlineStatus.upsert({
        where: { userId },
        update: {
          username,
          customStatus: dto.customStatus,
          isPrivate: dto.isPrivate,
        },
        create: {
          userId,
          username,
          customStatus: dto.customStatus,
          isPrivate: dto.isPrivate ?? false,
        },
      });
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(
        `Failed to update custom status for user ${userId}`,
        err,
      );
      throw err;
    }
  }

  async getUserStatus(userId: number): Promise<UserStatus | null> {
    try {
      const [onlineStatus, onlineKey, lastSeen] = await Promise.all([
        this.prisma.onlineStatus.findUnique({ where: { userId } }),
        this.redis.get(`online-status:user:${userId}`),
        this.redis.get(`last-seen:user:${userId}`),
      ]);

      if (!onlineStatus) return null;

      const isOnline = !!onlineKey;
      const lastSeenTimestamp = lastSeen ? parseInt(lastSeen, 10) : null;

      return this.buildUserStatus(onlineStatus, isOnline, lastSeenTimestamp);
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to get status for user ${userId}`, err);
      return null;
    }
  }

  async getUserStatusByUsername(username: string): Promise<UserStatus | null> {
    try {
      const onlineStatus = await this.prisma.onlineStatus.findFirst({
        where: { username },
      });

      if (!onlineStatus) return null;

      const [onlineKey, lastSeen] = await Promise.all([
        this.redis.get(`online-status:user:${onlineStatus.userId}`),
        this.redis.get(`last-seen:user:${onlineStatus.userId}`),
      ]);

      const isOnline = !!onlineKey;
      const lastSeenTimestamp = lastSeen ? parseInt(lastSeen, 10) : null;

      return this.buildUserStatus(onlineStatus, isOnline, lastSeenTimestamp);
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to get status for username ${username}`, err);
      return null;
    }
  }

  async getBatchStatus(usernames: string[]): Promise<UserStatus[]> {
    try {
      const statuses = await this.prisma.onlineStatus.findMany({
        where: { username: { in: usernames } },
      });

      if (statuses.length === 0) return [];

      const redisKeys = statuses.flatMap((s) => [
        `online-status:user:${s.userId}`,
        `last-seen:user:${s.userId}`,
      ]);

      const redisValues = await this.redis.mget(redisKeys);

      const results: UserStatus[] = [];
      for (let i = 0; i < statuses.length; i++) {
        const onlineStatus = statuses[i];
        const onlineKey = redisValues[i * 2];
        const lastSeen = redisValues[i * 2 + 1];

        const isOnline = !!onlineKey;
        const lastSeenTimestamp = lastSeen ? parseInt(lastSeen, 10) : null;

        results.push(
          this.buildUserStatus(onlineStatus, isOnline, lastSeenTimestamp),
        );
      }

      return results;
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to get batch status`, err);
      return [];
    }
  }

  private buildUserStatus(
    onlineStatus: OnlineStatus,
    isOnline: boolean,
    lastSeen: number | null,
  ): UserStatus {
    return {
      userId: onlineStatus.userId,
      username: onlineStatus.username,
      customStatus: onlineStatus.customStatus,
      isPrivate: onlineStatus.isPrivate,
      isOnline,
      lastSeen: onlineStatus.isPrivate ? null : lastSeen,
    };
  }

  // === Centrifugo presence handling ===

  async handleUserPresence(userId: number, isOnline: boolean): Promise<void> {
    try {
      if (isOnline) {
        await this.markOnline(userId);
      } else {
        await this.markOffline(userId);
      }
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));
      this.logger.error(`Failed to handle presence for user ${userId}`, err);
    }
  }
}

import { Injectable, Logger, BadRequestException } from '@nestjs/common';
import { RedisService, RedisPipeline } from 'src/shared/redis/redis.service';
import { UpdateStatusDto } from './dto/update-status.dto';
import { UpsertStatusDto } from './dto/upsert-status.dto';
import { PrismaService } from '../shared/prisma/prisma.service';
import { OnlineStatus } from '@prisma/client';
import { UserStatus } from './schemas/user-status.schema';

const LAST_SEEN_TTL = 90 * 24 * 3600; // 90 дней

@Injectable()
export class StatusService {
  private readonly logger = new Logger(StatusService.name);

  constructor(
    private readonly prisma: PrismaService,
    private readonly redis: RedisService,
  ) {}

  // === NATS Event Handler ===

  async createUserStatusRecord(userData: {
    userId: number;
    username: string;
    email?: string;
  }): Promise<void> {
    try {
      // Валидация username
      if (userData.username.length > 32) {
        throw new BadRequestException(
          'Username must be less than 32 characters',
        );
      }

      await this.prisma.onlineStatus.upsert({
        where: { userId: userData.userId },
        update: {
          username: userData.username,
        },
        create: {
          userId: userData.userId,
          username: userData.username,
          customStatus: null,
          isPrivate: false,
        },
      });

      this.logger.log(`Created status record for user: ${userData.username}`);
    } catch (error) {
      this.logger.error(
        `Failed to create status record for user ${userData.userId}`,
        error,
      );
      throw error;
    }
  }

  // === Online status (Redis) ===

  async markOnline(userId: number): Promise<void> {
    try {
      const pipeline: RedisPipeline = this.redis.multi();
      pipeline.set(`online-status:user:${userId}`, '1', { EX: 70 });
      pipeline.set(`last-seen:user:${userId}`, Date.now().toString(), {
<<<<<<< Updated upstream
        EX: LAST_SEEN_TTL,
=======
        EX: 30 * 24 * 3600 * 12,
>>>>>>> Stashed changes
      });
      await pipeline.exec();
    } catch (error) {
      this.logger.error(`Failed to mark user ${userId} as online`, error);
    }
  }

  async markOffline(userId: number): Promise<void> {
    try {
      const pipeline: RedisPipeline = this.redis.multi();
      pipeline.del(`online-status:user:${userId}`);
      pipeline.set(`last-seen:user:${userId}`, Date.now().toString(), {
        EX: LAST_SEEN_TTL,
      });
      await pipeline.exec();
    } catch (error) {
      this.logger.error(`Failed to mark user ${userId} as offline`, error);
    }
  }

  async isOnline(userId: number): Promise<boolean> {
    try {
      return await this.redis.exists(`online-status:user:${userId}`);
    } catch (error) {
      this.logger.error(
        `Failed to check online status for user ${userId}`,
        error,
      );
      return false;
    }
  }

  // === Custom status (PostgreSQL) ===
  async updateUsername(userId: number, newUsername: string): Promise<void> {
    try {
      if (newUsername.length > 32) {
        throw new BadRequestException(
          'Username must be less than 32 characters',
        );
      }

      const existing = await this.prisma.onlineStatus.findUnique({
        where: { userId },
      });

      if (!existing) {
        // Если записи нет, создаем новую
        await this.prisma.onlineStatus.create({
          data: {
            userId,
            username: newUsername,
            customStatus: null,
            isPrivate: false,
          },
        });
      } else {
        // Обновляем существующую
        await this.prisma.onlineStatus.update({
          where: { userId },
          data: { username: newUsername },
        });
      }

      this.logger.log(`Updated username for user ${userId} to ${newUsername}`);
    } catch (error) {
      this.logger.error(`Failed to update username for user ${userId}`, error);
      throw error;
    }
  }

  async updateCustomStatus(
    userId: number,
    username: string,
    dto: UpsertStatusDto,
  ): Promise<void> {
    try {
      // Валидация username
      if (username.length > 32) {
        throw new BadRequestException(
          'Username must be less than 32 characters',
        );
      }

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
      this.logger.error(
        `Failed to update custom status for user ${userId}`,
        error,
      );
      throw error;
    }
  }

  async getUserStatus(
    userId: number,
  ): Promise<Omit<UserStatus, 'userId'> | null> {
    try {
      const [onlineStatus, onlineKey, lastSeen] = await Promise.all([
        this.prisma.onlineStatus.findUnique({ where: { userId } }),
        this.redis.get(`online-status:user:${userId}`),
        this.redis.get(`last-seen:user:${userId}`),
      ]);

      if (!onlineStatus) return null;

      const isOnline = !!onlineKey;
      const lastSeenTimestamp = lastSeen ? parseInt(lastSeen, 10) : null;

      return this.buildUserStatusResponse(
        onlineStatus,
        isOnline,
        lastSeenTimestamp,
      );
    } catch (error) {
      this.logger.error(`Failed to get status for user ${userId}`, error);
      return null;
    }
  }

  async getUserStatusByUsername(
    username: string,
  ): Promise<Omit<UserStatus, 'userId'> | null> {
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

      return this.buildUserStatusResponse(
        onlineStatus,
        isOnline,
        lastSeenTimestamp,
      );
    } catch (error) {
      this.logger.error(`Failed to get status for username ${username}`, error);
      return null;
    }
  }

  async getBatchStatus(
    usernames: string[],
  ): Promise<Omit<UserStatus, 'userId'>[]> {
    try {
      // Валидация длины usernames
      if (usernames.some((username) => username.length > 32)) {
        throw new BadRequestException(
          'All usernames must be less than 32 characters',
        );
      }

      const statuses = await this.prisma.onlineStatus.findMany({
        where: { username: { in: usernames } },
      });

      if (statuses.length === 0) return [];

      const redisKeys = statuses.flatMap((s) => [
        `online-status:user:${s.userId}`,
        `last-seen:user:${s.userId}`,
      ]);

      const redisValues = await this.redis.mget(redisKeys);

      const results: Omit<UserStatus, 'userId'>[] = [];
      for (let i = 0; i < statuses.length; i++) {
        const onlineStatus = statuses[i];
        const onlineKey = redisValues[i * 2];
        const lastSeen = redisValues[i * 2 + 1];

        const isOnline = !!onlineKey;
        const lastSeenTimestamp = lastSeen ? parseInt(lastSeen, 10) : null;

        results.push(
          this.buildUserStatusResponse(
            onlineStatus,
            isOnline,
            lastSeenTimestamp,
          ),
        );
      }

      return results;
    } catch (error) {
      this.logger.error(`Failed to get batch status`, error);
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

  private buildUserStatusResponse(
    onlineStatus: OnlineStatus,
    isOnline: boolean,
    lastSeen: number | null,
  ): Omit<UserStatus, 'userId'> {
    const status = this.buildUserStatus(onlineStatus, isOnline, lastSeen);
    const { userId, ...response } = status;
    return response;
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
      this.logger.error(`Failed to handle presence for user ${userId}`, error);
    }
  }
}

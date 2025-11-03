// src/status/nats.consumer.ts
import {
  Injectable,
  Logger,
  OnModuleInit,
  OnModuleDestroy,
} from '@nestjs/common';
import { UserRegisteredEventDto } from './dto/nats-event.dto';
import { StatusService } from './status.service';
import { NatsService } from '../shared/nats/nats.module';

@Injectable()
export class NatsConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsConsumer.name);
  private subscription: any = null;

  constructor(
    private readonly natsService: NatsService,
    private readonly statusService: StatusService,
  ) {}

  async onModuleInit() {
    await this.startConsumer();
  }

  async onModuleDestroy() {
    await this.stopConsumer();
  }

  private async startConsumer() {
    try {
      const js = this.natsService.getJetStreamClient();

      // Создаем/проверяем stream
      const jsm = await js.jetstreamManager();
      try {
        await jsm.streams.add({
          name: 'user-events-stream',
          subjects: ['user.*'],
          retention: 'limits',
          max_msgs: 10000,
          max_age: 24 * 60 * 60 * 1000, // 24 часа
        });
      } catch (error) {
        // Stream уже существует - это нормально
        if (!error.message.includes('stream name already in use')) {
          throw error;
        }
      }

      // Подписываемся на wildcard subject
      this.subscription = await js.subscribe('user.*', {
        config: {
          durable_name: 'status-service-consumer', // Уникальное имя для этого сервиса
          deliver_policy: 'all', // Получаем все сообщения
          ack_wait: 30 * 1000, // 30 секунд
        },
      });

      this.logger.log('NATS consumer started for user.* events');

      // Обработка сообщений
      this.processMessages();
    } catch (error) {
      this.logger.error('Failed to start NATS consumer', error);
    }
  }

  private async processMessages() {
    for await (const message of this.subscription) {
      try {
        await this.handleMessage(message);
      } catch (error) {
        this.logger.error('Error processing message', error);
        // В случае ошибки отрицательно подтверждаем сообщение
        message.nak();
      }
    }
  }

  private async handleMessage(message: any) {
    const data = this.natsService.getJsonCodec().decode(message.data);

    // Валидация данных
    const event = this.validateEvent(data);
    if (!event) {
      this.logger.warn('Invalid event data received', { data });
      message.term(); // окончательно отклоняем невалидное сообщение
      return;
    }

    this.logger.debug('Received NATS event', {
      subject: message.subject,
      type: event.type,
      username: event.username,
      userId: event.user_id,
    });

    // Обработка в зависимости от типа события
    switch (message.subject) {
      case 'user.registered':
        await this.handleUserRegistered(event);
        break;

      case 'user.code.resend':
        // Для событий повторной отправки кода можем просто логировать
        this.logger.debug('Code resend event received', {
          userId: event.user_id,
          email: event.email,
        });
        break;

      default:
        this.logger.warn('Unknown subject received', {
          subject: message.subject,
        });
        break;
    }

    // Подтверждаем обработку
    message.ack();
  }

  private validateEvent(data: any): UserRegisteredEventDto | null {
    try {
      // Простая валидация обязательных полей
      if (!data || typeof data !== 'object') return null;
      if (typeof data.type !== 'number') return null;
      if (typeof data.user_id !== 'number') return null;
      if (typeof data.username !== 'string' || !data.username.trim())
        return null;
      if (typeof data.email !== 'string' || !data.email.trim()) return null;

      return {
        type: data.type,
        code: data.code || 0,
        email: data.email,
        username: data.username,
        user_id: data.user_id,
      };
    } catch {
      return null;
    }
  }

  private async handleUserRegistered(event: UserRegisteredEventDto) {
    try {
      // При регистрации пользователя создаем запись в статусах
      await this.statusService.updateCustomStatus(
        event.user_id,
        event.username,
        {
          customStatus: 'Новый пользователь', // дефолтный статус
          isPrivate: false, // по умолчанию публичный
        },
      );

      this.logger.log('User status record created', {
        userId: event.user_id,
        username: event.username,
      });
    } catch (error) {
      this.logger.error('Failed to create user status record', {
        userId: event.user_id,
        username: event.username,
        error: error.message,
      });
      throw error; // Пробрасываем ошибку для nak()
    }
  }

  private async stopConsumer() {
    if (this.subscription) {
      await this.subscription.destroy();
      this.logger.log('NATS consumer stopped');
    }
  }
}

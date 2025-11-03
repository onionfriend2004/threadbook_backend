import { CentrifugoWebhookDto } from './dto/centrifugo-webhook.dto';
import { UpdateStatusDto } from './dto/update-status.dto';
import { BatchStatusDto } from './dto/batch-status.dto';
import { StatusService } from './status.service';
import { AuthGuard } from '../auth/auth.guard';
import { User } from '../auth/decorators/user.decorator';
import {
  BadRequestException,
  NotFoundException,
  Controller,
  UseGuards,
  Headers,
  Param,
  Patch,
  Post,
  Body,
  Get,
} from '@nestjs/common';
import * as crypto from 'crypto';

@Controller('api')
export class StatusController {
  constructor(private statusService: StatusService) {}

  @UseGuards(AuthGuard)
  @Post('user/status')
  async updateStatus(
    @User() user: { userId: number; username: string },
    @Body() dto: UpdateStatusDto,
  ) {
    await this.statusService.updateCustomStatus(
      user.userId,
      user.username,
      dto,
    );
    return { success: true };
  }

  @UseGuards(AuthGuard)
  @Get('user/status')
  async getOwnStatus(@User() user: { userId: number; username: string }) {
    return this.statusService.getUserStatus(user.userId);
  }

  @UseGuards(AuthGuard)
  @Get('user/:username/status')
  async getStatusByUsername(@Param('username') username: string) {
    const status = await this.statusService.getUserStatusByUsername(username);
    if (!status) {
      throw new NotFoundException('User not found');
    }
    return status;
  }

  @UseGuards(AuthGuard)
  @Post('user/status/batch')
  async getBatchStatus(@Body() dto: BatchStatusDto) {
    return this.statusService.getBatchStatus(dto.usernames);
  }

  @UseGuards(AuthGuard)
  @Patch('user/status/privacy')
  async togglePrivacy(
    @User() user: { userId: number; username: string },
    @Body('isPrivate') isPrivate: boolean,
  ) {
    // Передаём username из сессии (даже если не меняется — на случай смены ника)
    await this.statusService.updateCustomStatus(user.userId, user.username, {
      isPrivate,
    });
    return { success: true };
  }

  // Вебхук от Centrifugo — БЕЗ GUARD!
  @Post('webhooks/centrifugo/presence')
  async centrifugoWebhook(
    @Body() rawBody: unknown,
    @Headers('x-webhook-signature') signature: string,
  ) {
    const secret = process.env.CENTRIFUGO_WEBHOOK_SECRET;
    if (!secret) {
      throw new Error('CENTRIFUGO_WEBHOOK_SECRET is not configured');
    }

    const expectedSignature = crypto
      .createHmac('sha256', secret)
      .update(JSON.stringify(rawBody))
      .digest('hex');

    if (signature !== expectedSignature) {
      throw new BadRequestException('Invalid webhook signature');
    }

    const parsedBody = CentrifugoWebhookDto.parse(rawBody);
    const { event, user } = parsedBody;

    const userId = Number(user);
    if (isNaN(userId) || userId <= 0) {
      throw new BadRequestException('Invalid user ID');
    }

    const isOnline = event === 'connect';
    await this.statusService.handleUserPresence(userId, isOnline);
    return { ok: true };
  }
}

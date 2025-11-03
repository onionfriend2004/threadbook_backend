import { FastifyRequest } from 'fastify';
import { AuthService } from './auth.service';
import {
  UnauthorizedException,
  ExecutionContext,
  CanActivate,
  Injectable,
} from '@nestjs/common';

@Injectable()
export class AuthGuard implements CanActivate {
  constructor(private authService: AuthService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<FastifyRequest>();
    const sid = request.cookies?.sid;

    if (!sid) {
      throw new UnauthorizedException('unauthorized: missing sid cookie');
    }

    const user = await this.authService.authenticate(sid);
    if (!user) {
      throw new UnauthorizedException('unauthorized: invalid session');
    }

    request.user = user;
    return true;
  }
}

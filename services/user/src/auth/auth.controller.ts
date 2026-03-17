import {
  Controller,
  Post,
  Get,
  Body,
  UseGuards,
  Request,
  Res,
  UnauthorizedException,
} from '@nestjs/common';
import { Throttle } from '@nestjs/throttler';
import type { Response } from 'express';
import { AuthService } from './auth.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { JwtAuthGuard } from './jwt-auth.guard';

interface RequestWithCookies {
  cookies: { ACCESS_TOKEN?: string; REFRESH_TOKEN?: string };
}

const IS_PRODUCTION = process.env.NODE_ENV === 'production';

const ACCESS_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: IS_PRODUCTION,
  sameSite: 'lax' as const,
  maxAge: 15 * 60 * 1000,
  path: '/',
};

const REFRESH_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: IS_PRODUCTION,
  sameSite: 'lax' as const,
  maxAge: 7 * 24 * 60 * 60 * 1000,
  path: '/',
};

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Throttle({ default: { limit: 3, ttl: 60000 } })
  @Post('register')
  async register(
    @Body() dto: RegisterDto,
    @Res({ passthrough: true }) res: Response,
  ) {
    const { ACCESS_TOKEN, refresh_token } =
      await this.authService.register(dto);
    res.cookie('ACCESS_TOKEN', ACCESS_TOKEN, ACCESS_COOKIE_OPTIONS);
    res.cookie('REFRESH_TOKEN', refresh_token, REFRESH_COOKIE_OPTIONS);
    return { message: 'registered' };
  }

  @Throttle({ default: { limit: 5, ttl: 60000 } })
  @Post('login')
  async login(
    @Body() dto: LoginDto,
    @Res({ passthrough: true }) res: Response,
  ) {
    console.log('[USER-SERVICE] ===== LOGIN STARTED =====');
    const { ACCESS_TOKEN, refresh_token } = await this.authService.login(dto);
    
    console.log('[USER-SERVICE] Envoi des cookies:');
    console.log('[USER-SERVICE] - ACCESS_TOKEN length:', ACCESS_TOKEN.length);
    console.log('[USER-SERVICE] - REFRESH_TOKEN length:', refresh_token.length);
    console.log('[USER-SERVICE] Cookie options:', { path: '/', domain: 'localhost' });
    
    res.cookie('ACCESS_TOKEN', ACCESS_TOKEN, ACCESS_COOKIE_OPTIONS);
    res.cookie('REFRESH_TOKEN', refresh_token, REFRESH_COOKIE_OPTIONS);
    
    console.log('[USER-SERVICE] ===== LOGIN COMPLETED =====\n');
    return { message: 'logged in' };
  }

  @Post('refresh')
  async refresh(
    @Request() req: RequestWithCookies,
    @Res({ passthrough: true }) res: Response,
  ) {
    const refreshToken = req.cookies.REFRESH_TOKEN;
    if (!refreshToken) {
      throw new UnauthorizedException('No refresh token provided');
    }

    const { ACCESS_TOKEN } =
      await this.authService.refreshAccessToken(refreshToken);
    res.cookie('ACCESS_TOKEN', ACCESS_TOKEN, ACCESS_COOKIE_OPTIONS);
    return { message: 'token refreshed' };
  }

  @Post('logout')
  async logout(
    @Request() req: RequestWithCookies,
    @Res({ passthrough: true }) res: Response,
  ) {
    const refreshToken = req.cookies.REFRESH_TOKEN;
    await this.authService.logout(refreshToken);
    res.clearCookie('ACCESS_TOKEN');
    res.clearCookie('REFRESH_TOKEN');
    return { message: 'logged out' };
  }

  @UseGuards(JwtAuthGuard)
  @Get('me')
  getProfile(@Request() req: { user: { userId: string } }) {
    return this.authService.getProfile(req.user.userId);
  }
}

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
};

const REFRESH_COOKIE_OPTIONS = {
  httpOnly: true,
  secure: IS_PRODUCTION,
  sameSite: 'lax' as const,
  maxAge: 7 * 24 * 60 * 60 * 1000,
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
    const { access_token, refresh_token } =
      await this.authService.register(dto);
    res.cookie('ACCESS_TOKEN', access_token, ACCESS_COOKIE_OPTIONS);
    res.cookie('REFRESH_TOKEN', refresh_token, REFRESH_COOKIE_OPTIONS);
    return { message: 'registered' };
  }

  @Throttle({ default: { limit: 5, ttl: 60000 } })
  @Post('login')
  async login(
    @Body() dto: LoginDto,
    @Res({ passthrough: true }) res: Response,
  ) {
    const { access_token, refresh_token } = await this.authService.login(dto);
    res.cookie('ACCESS_TOKEN', access_token, ACCESS_COOKIE_OPTIONS);
    res.cookie('REFRESH_TOKEN', refresh_token, REFRESH_COOKIE_OPTIONS);
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

    const { access_token } =
      await this.authService.refreshAccessToken(refreshToken);
    res.cookie('ACCESS_TOKEN', access_token, ACCESS_COOKIE_OPTIONS);
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

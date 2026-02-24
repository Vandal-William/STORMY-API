import {
  Controller,
  Post,
  Get,
  Body,
  UseGuards,
  Request,
  Res,
} from '@nestjs/common';
import express from 'express';
import { AuthService } from './auth.service';
import { RegisterDto } from './dto/register.dto';
import { LoginDto } from './dto/login.dto';
import { JwtAuthGuard } from './jwt-auth.guard';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('register')
  async register(
    @Body() dto: RegisterDto,
    @Res({ passthrough: true }) res: express.Response,
  ) {
    const { access_token } = await this.authService.register(dto);
    res.cookie('ACCESS_TOKEN', access_token, {
      httpOnly: true,
      secure: false,
      sameSite: 'lax',
      maxAge: 24 * 60 * 60 * 1000,
    });
    return { message: 'registered' };
  }

  @Post('login')
  async login(
    @Body() dto: LoginDto,
    @Res({ passthrough: true }) res: express.Response,
  ) {
    const { access_token } = await this.authService.login(dto);
    res.cookie('ACCESS_TOKEN', access_token, {
      httpOnly: true,
      secure: false,
      sameSite: 'lax',
      maxAge: 24 * 60 * 60 * 1000,
    });
    return { message: 'logged in' };
  }

  @Post('logout')
  logout(@Res({ passthrough: true }) res: express.Response) {
    res.clearCookie('ACCESS_TOKEN');
    return { message: 'logged out' };
  }

  @UseGuards(JwtAuthGuard)
  @Get('me')
  getProfile(@Request() req: { user: { userId: string } }) {
    return this.authService.getProfile(req.user.userId);
  }
}

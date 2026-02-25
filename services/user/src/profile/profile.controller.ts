import {
  Controller,
  Get,
  Patch,
  Delete,
  Body,
  Param,
  Query,
  Request,
  UseGuards,
  ParseUUIDPipe,
  BadRequestException,
} from '@nestjs/common';
import { ProfileService } from './profile.service';
import { UpdateProfileDto } from './dto/update-profile.dto';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@UseGuards(JwtAuthGuard)
@Controller('profile')
export class ProfileController {
  constructor(private readonly profileService: ProfileService) {}

  @Get('me')
  async getOwnProfile(@Request() req) {
    return this.profileService.getOwnProfile(req.user.userId);
  }

  @Get('search')
  async searchUsers(
    @Query('username') username: string,
    @Query('page') page?: string,
    @Query('limit') limit?: string,
  ) {
    if (!username || username.trim().length === 0) {
      throw new BadRequestException(
        "Le paramètre 'username' est requis pour la recherche",
      );
    }

    const pageNum = Math.max(1, Number(page) || 1);
    const limitNum = Math.min(100, Math.max(1, Number(limit) || 20));

    return this.profileService.searchByUsername(
      username.trim(),
      pageNum,
      limitNum,
    );
  }

  @Get(':id')
  async getUserById(@Param('id', ParseUUIDPipe) id: string) {
    return this.profileService.getUserById(id);
  }

  @Patch('me')
  async updateProfile(@Request() req, @Body() dto: UpdateProfileDto) {
    return this.profileService.updateProfile(req.user.userId, dto);
  }

  @Delete('me')
  async deleteAccount(@Request() req) {
    return this.profileService.deleteAccount(req.user.userId);
  }
}

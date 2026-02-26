import {
  Controller,
  Post,
  Get,
  Delete,
  Body,
  Param,
  Query,
  ParseUUIDPipe,
} from '@nestjs/common';
import { UserBanService } from './user-ban.service';
import { CreateBanDto } from './dto/create-ban.dto';

@Controller('bans')
export class UserBanController {
  constructor(private readonly userBanService: UserBanService) {}

  @Post()
  async create(@Body() dto: CreateBanDto) {
    return this.userBanService.create(dto);
  }

  @Get()
  async findAll(
    @Query('page') page?: string,
    @Query('limit') limit?: string,
    @Query('active') active?: string,
  ) {
    const pageNum = Math.max(1, Number(page) || 1);
    const limitNum = Math.min(100, Math.max(1, Number(limit) || 20));

    return this.userBanService.findAll(pageNum, limitNum, active);
  }

  @Get('check/:userId')
  async checkBan(@Param('userId', ParseUUIDPipe) userId: string) {
    return this.userBanService.checkBan(userId);
  }

  @Get(':id')
  async findOne(@Param('id', ParseUUIDPipe) id: string) {
    return this.userBanService.findOne(id);
  }

  @Delete(':id')
  async remove(@Param('id', ParseUUIDPipe) id: string) {
    return this.userBanService.remove(id);
  }
}

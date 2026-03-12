import {
  Controller,
  Post,
  Get,
  Patch,
  Delete,
  Body,
  Param,
  Query,
  Request,
  UseGuards,
  ParseUUIDPipe,
} from '@nestjs/common';
import { ContactService } from './contact.service';
import {
  AddContactDto,
  UpdateContactDto,
  BlockUserDto,
} from './dto/contact.dto';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

interface AuthRequest {
  user: { userId: string; username: string };
}

@UseGuards(JwtAuthGuard)
@Controller('contacts')
export class ContactController {
  constructor(private readonly contactService: ContactService) {}

  // ========================
  // CONTACTS
  // ========================

  @Post()
  async addContact(@Request() req: AuthRequest, @Body() dto: AddContactDto) {
    return this.contactService.addContact(req.user.userId, dto);
  }

  @Get()
  async getContacts(
    @Request() req: AuthRequest,
    @Query('page') page?: string,
    @Query('limit') limit?: string,
  ) {
    const pageNum = Math.max(1, Number(page) || 1);
    const limitNum = Math.min(100, Math.max(1, Number(limit) || 20));
    return this.contactService.getContacts(req.user.userId, pageNum, limitNum);
  }

  @Patch(':id')
  async updateContact(
    @Request() req: AuthRequest,
    @Param('id', ParseUUIDPipe) id: string,
    @Body() dto: UpdateContactDto,
  ) {
    return this.contactService.updateContact(req.user.userId, id, dto);
  }

  @Delete(':id')
  async removeContact(
    @Request() req: AuthRequest,
    @Param('id', ParseUUIDPipe) id: string,
  ) {
    return this.contactService.removeContact(req.user.userId, id);
  }

  // ========================
  // BLOCKED USERS
  // ========================

  @Post('blocked')
  async blockUser(@Request() req: AuthRequest, @Body() dto: BlockUserDto) {
    return this.contactService.blockUser(req.user.userId, dto);
  }

  @Get('blocked')
  async getBlockedUsers(
    @Request() req: AuthRequest,
    @Query('page') page?: string,
    @Query('limit') limit?: string,
  ) {
    const pageNum = Math.max(1, Number(page) || 1);
    const limitNum = Math.min(100, Math.max(1, Number(limit) || 20));
    return this.contactService.getBlockedUsers(
      req.user.userId,
      pageNum,
      limitNum,
    );
  }

  @Delete('blocked/:id')
  async unblockUser(
    @Request() req: AuthRequest,
    @Param('id', ParseUUIDPipe) id: string,
  ) {
    return this.contactService.unblockUser(req.user.userId, id);
  }
}

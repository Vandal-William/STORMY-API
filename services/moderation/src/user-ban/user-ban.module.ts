import { Module } from '@nestjs/common';
import { UserBanController } from './user-ban.controller';
import { UserBanService } from './user-ban.service';

@Module({
  controllers: [UserBanController],
  providers: [UserBanService],
})
export class UserBanModule {}

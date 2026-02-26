import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { PrismaModule } from './prisma/prisma.module';
import { ReportModule } from './report/report.module';
import { UserBanModule } from './user-ban/user-ban.module';

@Module({
  imports: [PrismaModule, ReportModule, UserBanModule],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}

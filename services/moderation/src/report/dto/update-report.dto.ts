import { IsEnum } from 'class-validator';
import { ReportStatus } from '@prisma/client';

export class UpdateReportDto {
  @IsEnum(ReportStatus, {
    message: `Le statut doit être l'un des suivants : ${Object.values(ReportStatus).join(', ')}`,
  })
  status: ReportStatus;
}

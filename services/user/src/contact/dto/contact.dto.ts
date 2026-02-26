import {
  IsOptional,
  IsString,
  IsUUID,
  MaxLength,
} from 'class-validator';

export class AddContactDto {
  @IsUUID('4', { message: "L'identifiant du contact doit être un UUID valide" })
  contactUserId: string;

  @IsOptional()
  @IsString({ message: 'Le surnom doit être une chaîne de caractères' })
  @MaxLength(100, { message: 'Le surnom ne doit pas dépasser 100 caractères' })
  nickname?: string;
}

export class UpdateContactDto {
  @IsOptional()
  @IsString({ message: 'Le surnom doit être une chaîne de caractères' })
  @MaxLength(100, { message: 'Le surnom ne doit pas dépasser 100 caractères' })
  nickname?: string;
}

export class BlockUserDto {
  @IsUUID('4', {
    message: "L'identifiant de l'utilisateur doit être un UUID valide",
  })
  blockedUserId: string;
}

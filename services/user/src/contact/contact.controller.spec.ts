import { Test, TestingModule } from '@nestjs/testing';
import { ContactController } from './contact.controller';
import { ContactService } from './contact.service';

describe('ContactController', () => {
  let controller: ContactController;

  const mockContactService = {
    addContact: jest.fn(),
    getContacts: jest.fn(),
    updateContact: jest.fn(),
    removeContact: jest.fn(),
    blockUser: jest.fn(),
    getBlockedUsers: jest.fn(),
    unblockUser: jest.fn(),
  };

  const authReq = { user: { userId: 'user-uuid', username: 'testuser' } };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [ContactController],
      providers: [{ provide: ContactService, useValue: mockContactService }],
    }).compile();

    controller = module.get<ContactController>(ContactController);

    jest.clearAllMocks();
  });

  // ========================
  // CONTACTS
  // ========================

  describe('addContact', () => {
    it('should add a contact for the authenticated user', async () => {
      const dto = { contactUserId: 'contact-uuid', nickname: 'Ami' };
      const created = { id: 'entry-uuid', ...dto, userId: 'user-uuid' };
      mockContactService.addContact.mockResolvedValue(created);

      const result = await controller.addContact(authReq, dto);

      expect(result).toEqual(created);
      expect(mockContactService.addContact).toHaveBeenCalledWith('user-uuid', dto);
    });
  });

  describe('getContacts', () => {
    it('should return contacts with default pagination', async () => {
      const contacts = { data: [], total: 0, page: 1, limit: 20, totalPages: 0 };
      mockContactService.getContacts.mockResolvedValue(contacts);

      const result = await controller.getContacts(authReq);

      expect(result).toEqual(contacts);
      expect(mockContactService.getContacts).toHaveBeenCalledWith('user-uuid', 1, 20);
    });

    it('should pass custom page and limit', async () => {
      mockContactService.getContacts.mockResolvedValue({ data: [], total: 0, page: 2, limit: 10, totalPages: 0 });

      await controller.getContacts(authReq, '2', '10');

      expect(mockContactService.getContacts).toHaveBeenCalledWith('user-uuid', 2, 10);
    });

    it('should clamp page to minimum 1 and limit to max 100', async () => {
      mockContactService.getContacts.mockResolvedValue({ data: [], total: 0, page: 1, limit: 100, totalPages: 0 });

      await controller.getContacts(authReq, '-1', '999');

      expect(mockContactService.getContacts).toHaveBeenCalledWith('user-uuid', 1, 100);
    });
  });

  describe('updateContact', () => {
    it('should update a contact', async () => {
      const dto = { nickname: 'New name' };
      const updated = { id: 'entry-uuid', nickname: 'New name' };
      mockContactService.updateContact.mockResolvedValue(updated);

      const result = await controller.updateContact(authReq, 'entry-uuid', dto);

      expect(result).toEqual(updated);
      expect(mockContactService.updateContact).toHaveBeenCalledWith('user-uuid', 'entry-uuid', dto);
    });
  });

  describe('removeContact', () => {
    it('should remove a contact', async () => {
      mockContactService.removeContact.mockResolvedValue({ deleted: true });

      const result = await controller.removeContact(authReq, 'entry-uuid');

      expect(result).toEqual({ deleted: true });
      expect(mockContactService.removeContact).toHaveBeenCalledWith('user-uuid', 'entry-uuid');
    });
  });

  // ========================
  // BLOCKED USERS
  // ========================

  describe('blockUser', () => {
    it('should block a user', async () => {
      const dto = { blockedUserId: 'blocked-uuid' };
      const blocked = { id: 'blocked-entry-uuid', ...dto, userId: 'user-uuid' };
      mockContactService.blockUser.mockResolvedValue(blocked);

      const result = await controller.blockUser(authReq, dto);

      expect(result).toEqual(blocked);
      expect(mockContactService.blockUser).toHaveBeenCalledWith('user-uuid', dto);
    });
  });

  describe('getBlockedUsers', () => {
    it('should return blocked users with default pagination', async () => {
      const blocked = { data: [], total: 0, page: 1, limit: 20, totalPages: 0 };
      mockContactService.getBlockedUsers.mockResolvedValue(blocked);

      const result = await controller.getBlockedUsers(authReq);

      expect(result).toEqual(blocked);
      expect(mockContactService.getBlockedUsers).toHaveBeenCalledWith('user-uuid', 1, 20);
    });
  });

  describe('unblockUser', () => {
    it('should unblock a user', async () => {
      mockContactService.unblockUser.mockResolvedValue({ deleted: true });

      const result = await controller.unblockUser(authReq, 'blocked-entry-uuid');

      expect(result).toEqual({ deleted: true });
      expect(mockContactService.unblockUser).toHaveBeenCalledWith('user-uuid', 'blocked-entry-uuid');
    });
  });
});

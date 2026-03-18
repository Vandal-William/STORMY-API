import { Test, TestingModule } from '@nestjs/testing';
import { BadRequestException } from '@nestjs/common';
import { ProfileController } from './profile.controller';
import { ProfileService } from './profile.service';

describe('ProfileController', () => {
  let controller: ProfileController;

  const mockProfileService = {
    getOwnProfile: jest.fn(),
    getUserById: jest.fn(),
    searchByUsername: jest.fn(),
    updateProfile: jest.fn(),
    deleteAccount: jest.fn(),
  };

  const authReq = { user: { userId: 'user-uuid', username: 'testuser' } };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [ProfileController],
      providers: [{ provide: ProfileService, useValue: mockProfileService }],
    }).compile();

    controller = module.get<ProfileController>(ProfileController);

    jest.clearAllMocks();
  });

  describe('getOwnProfile', () => {
    it('should return the authenticated user profile', async () => {
      const profile = { id: 'user-uuid', username: 'testuser', phone: '0612345678' };
      mockProfileService.getOwnProfile.mockResolvedValue(profile);

      const result = await controller.getOwnProfile(authReq);

      expect(result).toEqual(profile);
      expect(mockProfileService.getOwnProfile).toHaveBeenCalledWith('user-uuid');
    });
  });

  describe('searchUsers', () => {
    it('should search users by username', async () => {
      const searchResult = {
        data: [{ id: '1', username: 'alice' }],
        total: 1,
        page: 1,
        limit: 20,
        totalPages: 1,
      };
      mockProfileService.searchByUsername.mockResolvedValue(searchResult);

      const result = await controller.searchUsers('alice');

      expect(result).toEqual(searchResult);
      expect(mockProfileService.searchByUsername).toHaveBeenCalledWith('alice', 1, 20);
    });

    it('should handle page and limit query params', async () => {
      mockProfileService.searchByUsername.mockResolvedValue({ data: [], total: 0, page: 2, limit: 10, totalPages: 0 });

      await controller.searchUsers('test', '2', '10');

      expect(mockProfileService.searchByUsername).toHaveBeenCalledWith('test', 2, 10);
    });

    it('should throw BadRequestException if username is empty', async () => {
      await expect(controller.searchUsers('')).rejects.toThrow(BadRequestException);
    });

    it('should throw BadRequestException if username is only spaces', async () => {
      await expect(controller.searchUsers('   ')).rejects.toThrow(BadRequestException);
    });

    it('should clamp page to minimum 1', async () => {
      mockProfileService.searchByUsername.mockResolvedValue({ data: [], total: 0, page: 1, limit: 20, totalPages: 0 });

      await controller.searchUsers('test', '-5');

      expect(mockProfileService.searchByUsername).toHaveBeenCalledWith('test', 1, 20);
    });

    it('should clamp limit to maximum 100', async () => {
      mockProfileService.searchByUsername.mockResolvedValue({ data: [], total: 0, page: 1, limit: 100, totalPages: 0 });

      await controller.searchUsers('test', '1', '500');

      expect(mockProfileService.searchByUsername).toHaveBeenCalledWith('test', 1, 100);
    });
  });

  describe('getUserById', () => {
    it('should return user public profile by id', async () => {
      const user = { id: 'other-uuid', username: 'otheruser' };
      mockProfileService.getUserById.mockResolvedValue(user);

      const result = await controller.getUserById('other-uuid');

      expect(result).toEqual(user);
      expect(mockProfileService.getUserById).toHaveBeenCalledWith('other-uuid');
    });
  });

  describe('updateProfile', () => {
    it('should update the authenticated user profile', async () => {
      const updatedProfile = { id: 'user-uuid', username: 'newname' };
      mockProfileService.updateProfile.mockResolvedValue(updatedProfile);

      const result = await controller.updateProfile(authReq, { username: 'newname' });

      expect(result).toEqual(updatedProfile);
      expect(mockProfileService.updateProfile).toHaveBeenCalledWith('user-uuid', { username: 'newname' });
    });
  });

  describe('deleteAccount', () => {
    it('should delete the authenticated user account', async () => {
      mockProfileService.deleteAccount.mockResolvedValue({ deleted: true });

      const result = await controller.deleteAccount(authReq);

      expect(result).toEqual({ deleted: true });
      expect(mockProfileService.deleteAccount).toHaveBeenCalledWith('user-uuid');
    });
  });
});

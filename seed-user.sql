-- Créer les tables pour le service User
-- Basé sur le schéma Prisma
-- kubectl exec -it svc/postgres -- psql -U postgres -d app -f /dev/stdin < /home/william/Stormy/seed-user.sql


-- Supprimer les tables si elles existent (pour éviter les conflits)
DROP TABLE IF EXISTS blocked_users CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Table users
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone VARCHAR(20) UNIQUE NOT NULL,
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(255),
  password_hash VARCHAR(255) NOT NULL,
  avatar_url VARCHAR(500),
  about TEXT,
  last_seen TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table contacts
CREATE TABLE IF NOT EXISTS contacts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  contact_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nickname VARCHAR(100),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, contact_user_id)
);

-- Table blocked_users
CREATE TABLE IF NOT EXISTS blocked_users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  blocked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, blocked_user_id)
);

-- Insérer 4 utilisateurs avec données fictives
-- Password: password123 hashé avec bcrypt
INSERT INTO users (id, phone, username, email, password_hash, avatar_url, about, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', '0612345678', 'alice_wonder', 'alice@example.com', '$2a$12$7o589Qe0cRD3yFg3Ho4QIObZxw3BhkTpP0/OhQVaPt8DU1WM6so6q', 'https://api.example.com/avatars/alice.jpg', 'Creative designer and tech enthusiast', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('550e8400-e29b-41d4-a716-446655440002', '0623456789', 'bob_builder', 'bob@example.com', '$2a$12$7o589Qe0cRD3yFg3Ho4QIObZxw3BhkTpP0/OhQVaPt8DU1WM6so6q', 'https://api.example.com/avatars/bob.jpg', 'Software engineer, coffee lover', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('550e8400-e29b-41d4-a716-446655440003', '0634567890', 'charlie_chess', 'charlie@example.com', '$2a$12$7o589Qe0cRD3yFg3Ho4QIObZxw3BhkTpP0/OhQVaPt8DU1WM6so6q', 'https://api.example.com/avatars/charlie.jpg', 'Chess player and data scientist', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
('550e8400-e29b-41d4-a716-446655440004', '0645678901', 'diana_dance', 'diana@example.com', '$2a$12$7o589Qe0cRD3yFg3Ho4QIObZxw3BhkTpP0/OhQVaPt8DU1WM6so6q', 'https://api.example.com/avatars/diana.jpg', 'Dancer and fitness coach', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Ajouter des contacts (relations d'amitié)
INSERT INTO contacts (id, user_id, contact_user_id, nickname, created_at) VALUES
('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', 'Bob', CURRENT_TIMESTAMP),
('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440003', 'Charlie', CURRENT_TIMESTAMP),
('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Alice', CURRENT_TIMESTAMP),
('660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440004', 'Diana', CURRENT_TIMESTAMP),
('660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440004', 'Diana', CURRENT_TIMESTAMP);

-- Ajouter un utilisateur bloqué
INSERT INTO blocked_users (id, user_id, blocked_user_id, created_at) VALUES
('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP);

-- Afficher les résultats
SELECT '===== UTILISATEURS CRÉÉS =====' as info;
SELECT id, username, email, phone FROM users;

SELECT '===== CONTACTS =====' as info;
SELECT c.id, u1.username as user_from, u2.username as user_to, c.nickname FROM contacts c
JOIN users u1 ON c.user_id = u1.id
JOIN users u2 ON c.contact_user_id = u2.id;

SELECT '===== UTILISATEURS BLOQUÉS =====' as info;
SELECT b.id, u1.username as blocker, u2.username as blocked FROM blocked_users b
JOIN users u1 ON b.user_id = u1.id
JOIN users u2 ON b.blocked_user_id = u2.id;

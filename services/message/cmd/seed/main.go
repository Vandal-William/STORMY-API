package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
)

// TestData represents sample data for testing
type TestData struct {
	Users         []TestUser
	Conversations []TestConversation
	Messages      []TestMessage
}

// TestUser represents a test user
type TestUser struct {
	ID int32
}

// TestConversation represents a test conversation
type TestConversation struct {
	ID        gocql.UUID
	CreatedBy int32
	Title     string
	Type      string // "direct" or "group"
}

// TestMessage represents a test message
type TestMessage struct {
	ID             gocql.UUID
	ConversationID gocql.UUID
	SenderID       int32
	Content        string
	CreatedAt      time.Time
}

// CassandraSeed initializes test data in Cassandra
type CassandraSeed struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

// NewCassandraSeed creates a new Cassandra seeder
func NewCassandraSeed(hosts []string, port int) *CassandraSeed {
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = port
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	return &CassandraSeed{
		cluster: cluster,
	}
}

// Connect establishes connection to Cassandra
func (cs *CassandraSeed) Connect() error {
	session, err := cs.cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("failed to connect to Cassandra: %w", err)
	}
	cs.session = session
	log.Println("✓ Connected to Cassandra")
	return nil
}

// Close closes the Cassandra connection
func (cs *CassandraSeed) Close() error {
	if cs.session != nil {
		cs.session.Close()
	}
	return nil
}

// GenerateTestData creates sample data
func GenerateTestData() *TestData {
	// Create UUIDs for conversations and messages
	conv1UUID := gocql.MustParseUUID("650e8400-e29b-41d4-a716-446655440001")
	conv2UUID := gocql.MustParseUUID("650e8400-e29b-41d4-a716-446655440002")
	conv3UUID := gocql.MustParseUUID("650e8400-e29b-41d4-a716-446655440003")

	msg1UUID := gocql.MustParseUUID("750e8400-e29b-41d4-a716-446655440001")
	msg2UUID := gocql.MustParseUUID("750e8400-e29b-41d4-a716-446655440002")
	msg3UUID := gocql.MustParseUUID("750e8400-e29b-41d4-a716-446655440003")
	msg4UUID := gocql.MustParseUUID("750e8400-e29b-41d4-a716-446655440004")
	msg5UUID := gocql.MustParseUUID("750e8400-e29b-41d4-a716-446655440005")

	// User IDs as integers (from user service)
	user1ID := int32(1)
	user2ID := int32(2)
	user3ID := int32(3)

	now := time.Now()

	data := &TestData{
		Users: []TestUser{
			{ID: user1ID},
			{ID: user2ID},
			{ID: user3ID},
		},
		Conversations: []TestConversation{
			{
				ID:        conv1UUID,
				CreatedBy: user1ID,
				Title:     "Alice & Bob Direct",
				Type:      "direct",
			},
			{
				ID:        conv2UUID,
				CreatedBy: user1ID,
				Title:     "Project Discussion",
				Type:      "group",
			},
			{
				ID:        conv3UUID,
				CreatedBy: user2ID,
				Title:     "Random Chat",
				Type:      "group",
			},
		},
		Messages: []TestMessage{
			{
				ID:             msg1UUID,
				ConversationID: conv1UUID,
				SenderID:       user1ID,
				Content:        "Hey Bob, how are you?",
				CreatedAt:      now.Add(-2 * time.Hour),
			},
			{
				ID:             msg2UUID,
				ConversationID: conv1UUID,
				SenderID:       user2ID,
				Content:        "Hi Alice! I'm doing great, thanks for asking!",
				CreatedAt:      now.Add(-1*time.Hour - 30*time.Minute),
			},
			{
				ID:             msg3UUID,
				ConversationID: conv2UUID,
				SenderID:       user1ID,
				Content:        "Team, please review the latest design specs",
				CreatedAt:      now.Add(-1 * time.Hour),
			},
			{
				ID:             msg4UUID,
				ConversationID: conv2UUID,
				SenderID:       user2ID,
				Content:        "Looks good! I'll update the frontend accordingly",
				CreatedAt:      now.Add(-30 * time.Minute),
			},
			{
				ID:             msg5UUID,
				ConversationID: conv3UUID,
				SenderID:       user3ID,
				Content:        "Anyone up for lunch later?",
				CreatedAt:      now.Add(-15 * time.Minute),
			},
		},
	}

	return data
}

// SeedConversations inserts conversation test data
func (cs *CassandraSeed) SeedConversations(convs []TestConversation) error {
	log.Println("Seeding conversations...")

	for _, conv := range convs {
		query := `
			INSERT INTO message_service.conversations 
			(id, created_by, title, type, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		now := time.Now()
		if err := cs.session.Query(query,
			conv.ID,
			conv.CreatedBy,
			conv.Title,
			conv.Type,
			now,
			now,
		).Exec(); err != nil {
			return fmt.Errorf("failed to insert conversation %s: %w", conv.ID, err)
		}
		log.Printf("  ✓ Created conversation: %s (%s)", conv.Title, conv.ID)
	}

	return nil
}

// SeedConversationMembers inserts conversation members
func (cs *CassandraSeed) SeedConversationMembers(data *TestData) error {
	log.Println("Seeding conversation members...")

	members := []struct {
		conversationID gocql.UUID
		userIDs        []int32
		role           string
	}{
		{
			conversationID: data.Conversations[0].ID, // Alice & Bob
			userIDs:        []int32{data.Users[0].ID, data.Users[1].ID},
			role:           "member",
		},
		{
			conversationID: data.Conversations[1].ID, // Project Discussion
			userIDs:        []int32{data.Users[0].ID, data.Users[1].ID, data.Users[2].ID},
			role:           "member",
		},
		{
			conversationID: data.Conversations[2].ID, // Random Chat
			userIDs:        []int32{data.Users[1].ID, data.Users[2].ID},
			role:           "member",
		},
	}

	for _, m := range members {
		for _, userID := range m.userIDs {
			query := `
				INSERT INTO message_service.conversation_members
				(conversation_id, user_id, role, joined_at)
				VALUES (?, ?, ?, ?)
			`
			if err := cs.session.Query(query,
				m.conversationID,
				userID,
				m.role,
				time.Now(),
			).Exec(); err != nil {
				return fmt.Errorf("failed to insert conversation member: %w", err)
			}
		}
		log.Printf("  ✓ Added %d members to conversation %s", len(m.userIDs), m.conversationID)
	}

	return nil
}

// SeedMessages inserts message test data
func (cs *CassandraSeed) SeedMessages(messages []TestMessage) error {
	log.Println("Seeding messages...")

	for _, msg := range messages {
		query := `
			INSERT INTO message_service.messages
			(id, conversation_id, sender_id, content, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`
		if err := cs.session.Query(query,
			msg.ID,
			msg.ConversationID,
			msg.SenderID,
			msg.Content,
			msg.CreatedAt,
			msg.CreatedAt,
		).Exec(); err != nil {
			return fmt.Errorf("failed to insert message %s: %w", msg.ID, err)
		}
		log.Printf("  ✓ Created message: %s (from user %d)", msg.ID, msg.SenderID)
	}

	return nil
}

// SeedUserConversations creates index entries for user conversations
func (cs *CassandraSeed) SeedUserConversations(data *TestData) error {
	log.Println("Seeding user conversation index...")

	// Map conversations to users
	userConversations := make(map[int32][]gocql.UUID)

	for _, member := range []struct {
		convID gocql.UUID
		userID int32
	}{
		{data.Conversations[0].ID, data.Users[0].ID},
		{data.Conversations[0].ID, data.Users[1].ID},
		{data.Conversations[1].ID, data.Users[0].ID},
		{data.Conversations[1].ID, data.Users[1].ID},
		{data.Conversations[1].ID, data.Users[2].ID},
		{data.Conversations[2].ID, data.Users[1].ID},
		{data.Conversations[2].ID, data.Users[2].ID},
	} {
		userConversations[member.userID] = append(userConversations[member.userID], member.convID)
	}

	for userID, convIDs := range userConversations {
		for _, convID := range convIDs {
			query := `
				INSERT INTO message_service.user_conversations
				(user_id, conversation_id, joined_at)
				VALUES (?, ?, ?)
			`
			if err := cs.session.Query(query,
				userID,
				convID,
				time.Now(),
			).Exec(); err != nil {
				return fmt.Errorf("failed to insert user conversation: %w", err)
			}
		}
		log.Printf("  ✓ Indexed %d conversations for user %d", len(convIDs), userID)
	}

	return nil
}

// VerifyData checks that data was inserted correctly
func (cs *CassandraSeed) VerifyData() error {
	log.Println("\nVerifying seeded data...")

	// Count conversations
	var convCount int
	if err := cs.session.Query("SELECT COUNT(*) FROM message_service.conversations").
		Scan(&convCount); err != nil {
		return fmt.Errorf("failed to count conversations: %w", err)
	}
	log.Printf("  ✓ Found %d conversations", convCount)

	// Count messages
	var msgCount int
	if err := cs.session.Query("SELECT COUNT(*) FROM message_service.messages").
		Scan(&msgCount); err != nil {
		return fmt.Errorf("failed to count messages: %w", err)
	}
	log.Printf("  ✓ Found %d messages", msgCount)

	// Count members
	var memberCount int
	if err := cs.session.Query("SELECT COUNT(*) FROM message_service.conversation_members").
		Scan(&memberCount); err != nil {
		return fmt.Errorf("failed to count members: %w", err)
	}
	log.Printf("  ✓ Found %d conversation members", memberCount)

	return nil
}

// Seed runs the complete seeding process
func (cs *CassandraSeed) Seed() error {
	// Generate test data
	data := GenerateTestData()

	log.Println("\n=== Cassandra Test Data Seeding ===")
	log.Printf("Ready to seed %d conversations and %d messages\n",
		len(data.Conversations), len(data.Messages))

	// Insert conversations
	if err := cs.SeedConversations(data.Conversations); err != nil {
		return err
	}

	// Insert conversation members
	if err := cs.SeedConversationMembers(data); err != nil {
		return err
	}

	// Insert messages
	if err := cs.SeedMessages(data.Messages); err != nil {
		return err
	}

	// Insert user-conversation index
	if err := cs.SeedUserConversations(data); err != nil {
		return err
	}

	// Verify
	if err := cs.VerifyData(); err != nil {
		return err
	}

	log.Println("\n✓ Seeding completed successfully!")
	log.Println("\nTest data created:")
	log.Println("  Users: 1, 2, 3 (from user service)")
	log.Println("  Conversations: 3 (1 direct, 2 groups)")
	log.Println("  Messages: 5 (spread across conversations)")
	log.Println("\nYou can now test with the gateway using:")
	log.Println("  curl -H 'Authorization: Bearer {token}' \\")
	log.Println("       http://gateway:8080/users/conversations")

	return nil
}

func main() {
	// Configuration from environment or defaults
	hosts := []string{"localhost"}
	if h := os.Getenv("CASSANDRA_HOSTS"); h != "" {
		hosts = []string{h}
	}

	port := 9042
	if p := os.Getenv("CASSANDRA_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	log.Printf("Connecting to Cassandra at %v:%d...", hosts, port)

	seeder := NewCassandraSeed(hosts, port)

	if err := seeder.Connect(); err != nil {
		log.Fatalf("✗ Connection failed: %v", err)
	}
	defer seeder.Close()

	if err := seeder.Seed(); err != nil {
		log.Fatalf("✗ Seeding failed: %v", err)
	}
}

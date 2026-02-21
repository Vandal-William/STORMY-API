package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"message-service/internal/config"
	"message-service/internal/domain"
	"message-service/internal/infrastructure/cassandra"
	"message-service/internal/repository"
)

// Example usage of Cassandra repositories
// Run with: go run example.go or include in tests

func main() {
	// Setup
	cfg := config.Load()
	cfg.Cassandra.Hosts = []string{"localhost"}
	cfg.Cassandra.Port = 9042

	// Connect to Cassandra
	client, err := cassandra.NewClient(cfg.Cassandra)
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Close()

	log.Println("✓ Connected to Cassandra")

	// Create repositories
	messageRepo := repository.NewCassandraMessageRepository(client)
	conversationRepo := repository.NewCassandraConversationRepository(client)

	ctx := context.Background()

	// Example 1: Create a conversation
	log.Println("\n=== Example 1: Create Conversation ===")
	conversation := &domain.Conversation{
		ID:          gocql.TimeUUID(),
		Type:        "private",
		Name:        "Chat with John",
		Description: "Personal conversation",
		AvatarURL:   "https://example.com/avatar.jpg",
		CreatedBy:   gocql.TimeUUID(),
	}

	createdConv, err := conversationRepo.Create(ctx, conversation)
	if err != nil {
		log.Printf("Error creating conversation: %v", err)
	} else {
		log.Printf("✓ Conversation created: %s", createdConv.ID)
	}

	// Example 2: Create messages
	log.Println("\n=== Example 2: Create Messages ===")
	for i := 0; i < 3; i++ {
		msg := &domain.Message{
			ID:             gocql.TimeUUID(),
			ConversationID: createdConv.ID,
			SenderID:       conversation.CreatedBy,
			Content:        fmt.Sprintf("Hello message %d", i+1),
			Type:           "text",
			CreatedAt:      time.Now(),
		}

		createdMsg, err := messageRepo.Create(ctx, msg)
		if err != nil {
			log.Printf("Error creating message: %v", err)
		} else {
			log.Printf("✓ Message created: %s", createdMsg.ID)
		}
	}

	// Example 3: Get message by ID
	log.Println("\n=== Example 3: Get Message by ID ===")
	/* Uncomment with a real message ID
	messageID := gocql.TimeUUID()  // Use actual ID from Example 2
	msg, err := messageRepo.GetByID(ctx, messageID)
	if err != nil {
		log.Printf("Error retrieving message: %v", err)
	} else if msg != nil {
		log.Printf("✓ Retrieved message: %s (Content: %s)", msg.ID, msg.Content)
	} else {
		log.Println("Message not found")
	}
	*/

	// Example 4: Get messages in conversation
	log.Println("\n=== Example 4: Get Conversation Messages ===")
	messages, _, err := messageRepo.GetByConversationID(ctx, createdConv.ID, 10, nil)
	if err != nil {
		log.Printf("Error retrieving messages: %v", err)
	} else {
		log.Printf("✓ Retrieved %d messages from conversation", len(messages))
		for _, msg := range messages {
			log.Printf("  - %s: %s", msg.CreatedAt.Format("15:04:05"), msg.Content)
		}
	}

	// Example 5: Add conversation member
	log.Println("\n=== Example 5: Add Conversation Member ===")
	newMemberID := gocql.TimeUUID()
	member := &domain.ConversationMember{
		ID:             gocql.TimeUUID(),
		ConversationID: createdConv.ID,
		UserID:         newMemberID,
		Role:           "member",
		IsMuted:        false,
		JoinedAt:       time.Now(),
	}

	err = conversationRepo.AddMember(ctx, member)
	if err != nil {
		log.Printf("Error adding member: %v", err)
	} else {
		log.Printf("✓ Member added to conversation")
	}

	// Example 6: Get conversation members
	log.Println("\n=== Example 6: Get Conversation Members ===")
	members, err := conversationRepo.GetMembers(ctx, createdConv.ID)
	if err != nil {
		log.Printf("Error retrieving members: %v", err)
	} else {
		log.Printf("✓ Retrieved %d members", len(members))
		for _, m := range members {
			log.Printf("  - User %s: Role=%s, JoinedAt=%s", m.UserID, m.Role, m.JoinedAt.Format("2006-01-02"))
		}
	}

	// Example 7: Update message
	log.Println("\n=== Example 7: Update Message ===")
	/* Uncomment with a real message ID
	updatedMsg := &domain.Message{
		Content: "Updated content",
	}
	msg, err := messageRepo.Update(ctx, messageID, updatedMsg)
	if err != nil {
		log.Printf("Error updating message: %v", err)
	} else {
		log.Printf("✓ Message updated: %s", msg.Content)
	}
	*/

	// Example 8: Add message attachment
	log.Println("\n=== Example 8: Add Message Attachment ===")
	// First create a message to attach to
	msgWithAttachment := &domain.Message{
		ID:             gocql.TimeUUID(),
		ConversationID: createdConv.ID,
		SenderID:       conversation.CreatedBy,
		Content:        "Check out this image",
		Type:           "image",
		CreatedAt:      time.Now(),
	}

	createdMsgWithFile, err := messageRepo.Create(ctx, msgWithAttachment)
	if err != nil {
		log.Printf("Error creating message: %v", err)
	} else {
		// Create attachment
		attachment := &domain.MessageAttachment{
			ID:           gocql.TimeUUID(),
			MessageID:    createdMsgWithFile.ID,
			FileURL:      "https://cdn.example.com/image.jpg",
			FileName:     "image.jpg",
			FileType:     "image/jpeg",
			FileSize:     512000,
			ThumbnailURL: "https://cdn.example.com/image_thumb.jpg",
		}

		err = messageRepo.CreateAttachment(ctx, attachment)
		if err != nil {
			log.Printf("Error creating attachment: %v", err)
		} else {
			log.Printf("✓ Attachment created: %s", attachment.FileName)
		}
	}

	// Example 9: Update conversation
	log.Println("\n=== Example 9: Update Conversation ===")
	updatedConv := &domain.Conversation{
		Name:        "Updated Chat Title",
		Description: "New description",
		AvatarURL:   "https://example.com/avatar-new.jpg",
	}

	result, err := conversationRepo.Update(ctx, createdConv.ID, updatedConv)
	if err != nil {
		log.Printf("Error updating conversation: %v", err)
	} else {
		log.Printf("✓ Conversation updated: %s", result.Name)
	}

	// Example 10: Get user's conversations
	log.Println("\n=== Example 10: Get User Conversations ===")
	userID := conversation.CreatedBy
	userConversations, err := conversationRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("Error retrieving user conversations: %v", err)
	} else {
		log.Printf("✓ User has %d conversations", len(userConversations))
		for _, conv := range userConversations {
			log.Printf("  - %s (Type: %s)", conv.Name, conv.Type)
		}
	}

	// Example 11: Soft delete message
	log.Println("\n=== Example 11: Delete Message (Soft Delete) ===")
	testMessageID := gocql.TimeUUID()
	err = messageRepo.Delete(ctx, testMessageID)
	if err != nil {
		log.Printf("Error deleting message: %v", err)
	} else {
		log.Printf("✓ Message marked as deleted (soft delete)")
	}

	// Example 12: Remove member from conversation
	log.Println("\n=== Example 12: Remove Conversation Member ===")
	err = conversationRepo.RemoveMember(ctx, createdConv.ID, newMemberID)
	if err != nil {
		log.Printf("Error removing member: %v", err)
	} else {
		log.Printf("✓ Member removed from conversation")
	}

	log.Println("\n✓ All examples completed!")
}

// Additional example: Batch operations
func exampleBatchOperations(client *cassandra.Client, ctx context.Context) {
	session := client.GetSession()

	// Create multiple messages in a batch (faster)
	batch := session.NewBatch(gocql.LoggedBatch)

	conversationID := gocql.TimeUUID()
	senderID := gocql.TimeUUID()

	for i := 0; i < 100; i++ {
		batch.Query(
			`INSERT INTO messages 
			(id, conversation_id, sender_id, content, type, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			gocql.TimeUUID(),
			conversationID,
			senderID,
			fmt.Sprintf("Batch message %d", i),
			"text",
			time.Now(),
		)
	}

	if err := session.ExecuteBatch(batch.WithContext(ctx)); err != nil {
		log.Printf("Batch insert error: %v", err)
	} else {
		log.Println("✓ Batch of 100 messages inserted")
	}
}

// Example: Check schema
func exampleCheckSchema(client *cassandra.Client) {
	session := client.GetSession()

	// List all tables in the keyspace
	query := `
	SELECT table_name 
	FROM system_schema.tables 
	WHERE keyspace_name = 'message_service'
	ORDER BY table_name`

	iter := session.Query(query).Iter()
	var tableName string

	log.Println("Tables in message_service keyspace:")
	for iter.Scan(&tableName) {
		log.Printf("  - %s", tableName)
	}

	if err := iter.Close(); err != nil {
		log.Printf("Error listing tables: %v", err)
	}
}

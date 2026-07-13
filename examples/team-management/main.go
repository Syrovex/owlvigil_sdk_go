package main

import (
	"context"
	"fmt"
	"log"

	owlvigil "github.com/owlvigil/owlvigil-go"
	"github.com/owlvigil/owlvigil-go/examples/internal/envfile"
	"github.com/owlvigil/owlvigil-go/management"
)

func main() {
	if err := envfile.Load(); err != nil {
		log.Fatal(err)
	}
	apiKey, err := envfile.Required("OWLVIGIL_API_KEY")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	workspaceID := int64(1) // Replace with your workspace ID

	// Initialize client with a service-account API key that has team and member scopes.
	client := management.NewClient(
		owlvigil.WithAPIKey(apiKey),
	)

	// Example 1: List teams
	fmt.Println("=== Teams ===")
	teams, _, err := client.ListTeams(ctx, workspaceID, management.ListOptions{Limit: 10})
	if err != nil {
		log.Fatalf("Failed to list teams: %v", err)
	}
	for _, team := range teams.Items {
		fmt.Printf("- %s (%d members)\n", team.Name, team.MemberCount)
	}

	// Example 2: List members
	fmt.Println("\n=== Members ===")
	members, _, err := client.ListMembers(ctx, workspaceID, management.ListOptions{Limit: 10})
	if err != nil {
		log.Fatalf("Failed to list members: %v", err)
	}
	for _, member := range members.Items {
		fmt.Printf("- %s (%s) - %s\n", member.Name, member.Email, member.Status)
	}

	// Example 3: Create team
	fmt.Println("\n=== Create Team ===")
	newTeam, _, err := client.CreateTeam(ctx, workspaceID, &management.CreateTeamRequest{
		Name:        "Engineering",
		Description: "Engineering team",
	})
	if err != nil {
		log.Printf("Failed to create team: %v", err)
	} else {
		fmt.Printf("Created team: %s (ID: %d)\n", newTeam.Name, newTeam.ID)
	}

	// Example 4: Invite member
	fmt.Println("\n=== Invite Member ===")
	invitation, _, err := client.CreateMember(ctx, workspaceID, &management.CreateMemberRequest{
		Email:   "newuser@example.com",
		RoleIDs: []int64{1}, // Replace with actual role ID
	})
	if err != nil {
		log.Printf("Failed to invite member: %v", err)
	} else {
		fmt.Printf("Invited: %s\n", invitation.Email)
	}
}

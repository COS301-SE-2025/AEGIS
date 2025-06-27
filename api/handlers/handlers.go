package handlers

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
	AdminHandler AdminInterface
	//AuthHandler     AuthInterface
	//CaseHandler     CaseInterface
	//EvidenceHandler EvidenceInterface
	//UserHandler     UserInterface
	//ThreadHandler   ThreadInterface
	//MessageHandler  MessageInterface
	//ChatHandler     ChatInterface
}

type AdminInterface interface {
	RegisterUser(c *gin.Context)   //create
	ListUsers(c *gin.Context)      //read
	UpdateUserRole(c *gin.Context) //is this all an admin can update?
	DeleteUser(c *gin.Context)     //delete
	//GetRoles(c *gin.Context)
}

type AuthInterface interface {
	Login(c *gin.Context) // Login
	Logout(c *gin.Context)
	ResetPassword(c *gin.Context)
	//RequestPasswordReset(c *gin.Context)
}

type CaseInterface interface {
	CreateCase(c *gin.Context) //CreateCase

	//GetCaseByID(c *gin.Context) //missing service function
	ListAllCases(c *gin.Context)      // GetAllCases
	ListFilteredCases(c *gin.Context) // GetFilteredCases
	ListCasesByUserID(c *gin.Context) // GetCasesByUser

	UpdateCaseStatus(c *gin.Context) // UpdateCaseStatus

	RemoveCollaborator(c *gin.Context) // RemoveCollaborator

	CreateCollaborator(c *gin.Context) // AssignUserToCase
	ListCollaborators(c *gin.Context)  // GetCollaborators
}

type EvidenceInterface interface {
	//UploadEvidence(c *gin.Context) //UNDER REVIEW

	ListEvidenceByCaseID(c *gin.Context)
	ListEvidenceByUserID(c *gin.Context)
	GetEvidenceByID(c *gin.Context)
	DownloadEvidenceByUserID(c *gin.Context)
	GetEvidenceMetadata(c *gin.Context)

	DeleteEvidenceByID(c *gin.Context)
}

type UserInterface interface {
	GetProfile(c *gin.Context)    // GetProfile
	UpdateProfile(c *gin.Context) // UpdateProfile
	GetUserRoles(c *gin.Context)  // GetUserRoles
}

type ThreadInterface interface {
	CreateThread(c *gin.Context)
	AddParticipant(c *gin.Context)

	GetThreadsByFileID(c *gin.Context)
	GetThreadsByCaseID(c *gin.Context)
	GetThreadParticipants(c *gin.Context)
	GetThreadByID(c *gin.Context)
	GetUserByID(c *gin.Context)

	UpdateThreadStatus(c *gin.Context)
	UpdateThreadPriority(c *gin.Context)
}

type MessageInterface interface {
	SendMessage(c *gin.Context)
	ApproveMessage(c *gin.Context)
	AddReaction(c *gin.Context)
	AddMentions(c *gin.Context)

	GetMessagesByThreadID(c *gin.Context)
	GetReplies(c *gin.Context)
	GetMessageByID(c *gin.Context)

	RemoveReaction(c *gin.Context)
}

type ChatInterface interface {
	CreateGroup(c *gin.Context)
	GetGroup(c *gin.Context)
	GetUserGroups(c *gin.Context)
	UpdateGroup(c *gin.Context)
	DeleteGroup(c *gin.Context)
	AddMemberToGroup(c *gin.Context)
	RemoveMemberFromGroup(c *gin.Context)
	IsUserInGroup(c *gin.Context)
	UpdateLastMessage(c *gin.Context)

	CreateMessage(ctx gin.Context)
	GetMessageByID(ctx gin.Context)
	GetMessages(ctx gin.Context)
	SearchMessages(ctx gin.Context)
	UpdateMessage(ctx gin.Context)
	DeleteMessage(ctx gin.Context)
	MarkMessagesAsRead(ctx gin.Context)
	GetUnreadCount(ctx gin.Context)

	GetGroupMembers(ctx gin.Context)
	IsGroupAdmin(ctx gin.Context)
}

func NewHandler(
	admin AdminInterface,
	//auth AuthInterface,
	//case_ CaseInterface,
	//evidence EvidenceInterface,
	//user UserInterface,
	//thread_ ThreadInterface,
	//message MessageInterface,
	//chat ChatInterface,

) *Handler {
	return &Handler{
		AdminHandler: admin,
		//AuthHandler:     auth,
		//CaseHandler:     case_,
		//EvidenceHandler: evidence,
		//UserHandler:     user,
		//ThreadHandler:   thread_,
		//MessageHandler:  message,
		//ChatHandler:     chat,
	}
}

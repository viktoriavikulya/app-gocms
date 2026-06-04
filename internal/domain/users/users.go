package users

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	Bio         string `json:"bio"`
	AvatarID    string `json:"avatar_media_id"`
	Active      bool   `json:"active"`
}

type AuthorProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Slug        string `json:"slug"`
	Bio         string `json:"bio"`
	AvatarID    string `json:"avatar_media_id"`
	Active      bool   `json:"active"`
}

func (u User) PublicAuthor() AuthorProfile {
	return AuthorProfile{ID: u.ID, DisplayName: u.DisplayName, Slug: u.Slug, Bio: u.Bio, AvatarID: u.AvatarID, Active: u.Active}
}

package users

// Flag represents user flags on Discord.
type Flag uint64

const (
	FlagStaff                 Flag = 1 << 0
	FlagPartner               Flag = 1 << 1
	FlagHypesquad             Flag = 1 << 2
	FlagBugHunterLevel1       Flag = 1 << 3
	FlagHypesquadOnlineHouse1 Flag = 1 << 6
	FlagHypesquadOnlineHouse2 Flag = 1 << 7
	FlagHypesquadOnlineHouse3 Flag = 1 << 8
	FlagPremiumEarlySupporter Flag = 1 << 9
	FlagTeamPseudoUser        Flag = 1 << 10
	FlagBugHunterLevel2       Flag = 1 << 14
	FlagVerifiedBot           Flag = 1 << 16
	FlagVerifiedDeveloper     Flag = 1 << 17
	FlagCertifiedModerator    Flag = 1 << 18
	FlagBotHTTPInteractions   Flag = 1 << 19
	FlagActiveDeveloper       Flag = 1 << 22
)

// PremiumType represents the type of Nitro subscription a user has.
type PremiumType int

const (
	PremiumTypeNone PremiumType = iota
	PremiumTypeNitroClassic
	PremiumTypeNitro
	PremiumTypeNitroBasic
)

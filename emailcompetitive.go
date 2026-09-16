package bird

type EmailCompetitiveService struct {
	resource
	Brands    *EmailCompetitiveBrandsService
	Watchlist *EmailCompetitiveWatchlistService
}

type EmailCompetitiveBrandsService struct {
	resource
}

type EmailCompetitiveWatchlistService struct {
	resource
	Brands *EmailCompetitiveWatchlistBrandsService
}

type EmailCompetitiveWatchlistBrandsService struct {
	resource
	Campaigns *EmailCompetitiveWatchlistBrandsCampaignsService
}

type EmailCompetitiveWatchlistBrandsCampaignsService struct {
	resource
}

package bird

type AssetsService struct {
	resource
	Folders    *AssetsFoldersService
	References *AssetsReferencesService
}

type AssetsFoldersService struct{ resource }
type AssetsReferencesService struct{ resource }

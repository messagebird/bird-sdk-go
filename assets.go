package bird

type AssetsService struct {
	resource
	Uploads    *AssetsUploadsService
	Folders    *AssetsFoldersService
	References *AssetsReferencesService
}

type AssetsFoldersService struct{ resource }
type AssetsReferencesService struct{ resource }

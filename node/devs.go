package node

// Special structure for the Developer rewards
type DevReward struct {
	DevGroup     string  // group or information about reward
	DevAddress   string  // FCT address where to send developer reward
	DevRewardPct float64 // % from total development rewards
}

// Hardcode developers who work active on PegNet to avoid manipulations from config files
//
// contains the percentage distribution of the total 2000
// defined in `pegnet` repo `modules/conversions/conversionlimit.go` as PerBlockDevelopers
//
// TODO: in v2.5 move into special Developer Rewards Chain (DRC)
var (
	DeveloperRewardAddreses = []DevReward{
		{"Listing Tech Support", "FA2i9WZqJnaKbJxDY2AZdVgewE28uCcSwoFt8LJCMtGCC7tpCa2n", 10.00},
		{"Architecture Dev for PegNet 2.5", "FA37cGXKWMtf2MmHy3n1rMCYeLVuR5MpDaP4VXVeFavjJCJLYYez", 17.0},
		{"Trading Bots Work", "FA2wDRieaBrWeZHVuXXWUHY6t9nKCVCCKAMS5xknLUExuVAq3ziS", 9.0},
		{"Mining Pool Support / Dev/ Infra Hosting/ Gateway Operation", "FA3LDEA5fcskV6ZoFpKE84qPcjd7GYjEnswGHMZXL1V9d14wmgh3", 9.0},
		{"Media Tech Support", "FA381EygeEXjZzB6hNvxbE4oSUzHZMfvGByMZoW5UrG1gHEKJcNK", 8.0},
		{"Social Coverage User Support", "FA2DxkaTx1k2oGfbTqvwVMScSHHac7JFRiBjRngjRnqQpeBxsLhA", 8.0},
		{"pTrader + PIPs", "FA2Ersb227gn7eWJ2HPsHZ5QqxfMBZhSjwixQ44dAS17CtRXSDRU", 8.0},
		{"Desktop Wallet + PIPs", "FA2eFEVUzTQZxNp3LYYgjPaaHUfGmuvShhtBdGB2BBWMeByPCmJy", 8.0},
		{"Explorer + Mobile", "FA2T72oxBxXvnujNdsVUshqFM2qV1W4nJy33nkrpxbYQV8rFbUPP", 5.0},
		{"Prosper / Staking GUI Upgrades", "FA2cEaq1GdGfFjhymiTEzW24DocZFZHNBqe9qkT18YPaL5ZzsgRi", 5.0},
		{"Payment Intergrations Work", "FA2YhZBZbc4V858ao7dJuAqRC4iwA3MrbZs7BHUPK7Mq19yYdMwZ", 3.0},
		{"General Development Tasks", "FA3PYuvrsDvkhnekokVNrgLn7JiL5pChSBTtR9gZB1mVGFVB7JRD", 3.0},
		{"Gateway Ethereum Upgrade", "FA2Wy7AzeoBuaXYnGu67xa5zdNkmqTbPryUgpy7qVPvj46GRZkep", 2.0},
		{"Statistics & Visualizations of PegNet", "FA3dsCiKGzwrTALfX4T2CKv8wCmNMxwJx3jS4jz1ST9fwge9Wrnm", 2.0},
		{"Trading Tech Support", "FA2a2nXgkBg7pL5wrgm99rLZDGFs2T8jfTgMuia6ep8ZMkVtPe8E", 3.00},
	}
)

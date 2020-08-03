package node

// Special structure
type DevReward struct {
	DevGroup     string  // group or comment about reward
	DevAddress   string  // address where to send developer reward
	DevRewardPct float64 // % from total development rewards
}

var (
	// Hardcode developers who work active on PegNet
	// to avoid manipulations from config files
	// TODO: v2.5 move into special Developer Rewards Chain (DRC)
	DeveloperRewardAddreses = []DevReward{
		{"Listing Tech Support", "FA2i9WZqJnaKbJxDY2AZdVgewE28uCcSwoFt8LJCMtGCC7tpCa2n", 10.00},
		{"Archecture Dev for PegNet 2.5", "FA37cGXKWMtf2MmHy3n1rMCYeLVuR5MpDaP4VXVeFavjJCJLYYez", 9.0},
		{"Trading Bots Work", "FA2wDRieaBrWeZHVuXXWUHY6t9nKCVCCKAMS5xknLUExuVAq3ziS", 9.0},
		{"Mining Pool Support / Dev/ Infra Hosting/ Gateway Operation", "FA3LDEA5fcskV6ZoFpKE84qPcjd7GYjEnswGHMZXL1V9d14wmgh3", 9.0},
		{"Media Tech Support", "FA381EygeEXjZzB6hNvxbE4oSUzHZMfvGByMZoW5UrG1gHEKJcNK", 8.0},
		{"Social Coverage User Support", "FA2DxkaTx1k2oGfbTqvwVMScSHHac7JFRiBjRngjRnqQpeBxsLhA", 8.0},
		{"pTrader + PIPs", "FA2Ersb227gn7eWJ2HPsHZ5QqxfMBZhSjwixQ44dAS17CtRXSDRU", 8.0},
		{"Desktop Wallet + PIPs", "	FA2eFEVUzTQZxNp3LYYgjPaaHUfGmuvShhtBdGB2BBWMeByPCmJy", 8.0},
		{"DeFi Integrations + PIPs Work", "FA2z6Nnaj8a5nXmwZD8tYhDENAx8ciVE3xeNeixKiFj22vZEZEdT", 8.0},
		{"Explorer + Mobile", "	FA2T72oxBxXvnujNdsVUshqFM2qV1W4nJy33nkrpxbYQV8rFbUPP", 5.0},
		{"Prosper / Staking GUI Upgrades", "FA2cEaq1GdGfFjhymiTEzW24DocZFZHNBqe9qkT18YPaL5ZzsgRi", 5.0},
		{"Payment Intergrations Work", "FA2YhZBZbc4V858ao7dJuAqRC4iwA3MrbZs7BHUPK7Mq19yYdMwZ", 3.0},
		{"General Development Tasks", "FA3PYuvrsDvkhnekokVNrgLn7JiL5pChSBTtR9gZB1mVGFVB7JRD", 3.0},
		{"Gateway Ethereum Upgrade", "FA2Wy7AzeoBuaXYnGu67xa5zdNkmqTbPryUgpy7qVPvj46GRZkep", 2.0},
		{"Statistics & Visualizations of PegNet", "FA3dsCiKGzwrTALfX4T2CKv8wCmNMxwJx3jS4jz1ST9fwge9Wrnm", 2.0},
		{"Trading Tech Support", "FA2a2nXgkBg7pL5wrgm99rLZDGFs2T8jfTgMuia6ep8ZMkVtPe8E", 3.00},
	}
)

// DevelopersPayouts for PIP16 sending rewards for developers
func (d *Pegnetd) DevelopersPayouts(tx *sql.Tx, fLog *log.Entry, height uint32, heightTimestamp time.Time) error {

	totalPayout := uint64(conversions.PerBlockDevelopers) * 144 // once a day, should be changed to SnapshotRate when it's integrated

	// we use hardcoded list of dev payouts
	for _, dev := range DeveloperRewardAddreses {

		// here PerBlockDevelopers is total payout value
		reward := (conversions.PerBlockDevelopers / 100) * dev.DevRewardPct

		// TODO: move real txid
		txid := fmt.Sprintf("%064d", height)

		fLog.WithFields(log.Fields{
			"total":     float64(totalPayout) / 1e8,
			"developer": len(dev.DevAddress),
			"PEG":       float64(reward) / 1e8, // Float is good enough here,
			"txid":      txid,
		}).Info("developer reward | paid out to")

	}

	return nil
}

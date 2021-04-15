package mappings

// AccountsCloned will hold the configuration for the cloned accounts index
var AccountsCloned = Object{
	"properties": Object{
		"delegationLegacyWaitingNum": Object{
			"type": "double",
		},
		"delegationLegacyActiveNum": Object{
			"type": "double",
		},
		"validatorsActiveNum": Object{
			"type": "double",
		},
		"validatorsTopUpNum": Object{
			"type": "double",
		},
		"delegationNum": Object{
			"type": "double",
		},
		"totalStakeNum": Object{
			"type": "double",
		},
		"totalBalanceWithStakeNum": Object{
			"type": "double",
		},
	},
}

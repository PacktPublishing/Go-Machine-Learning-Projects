package main

var datahints = []bool{
	false, // Id
	true,  // MSSubClass
	true,  // MSZoning
	false, // LotFrontage
	false, // LotArea
	true,  // Street
	true,  // Alley
	true,  // LotShape
	true,  // LandContour
	true,  // Utilities
	true,  // LotConfig
	true,  // LandSlope
	true,  // Neighborhood
	true,  // Condition1
	true,  // Condition2
	true,  // BldgType
	true,  // HouseStyle
	false, // OverallQual
	false, // OverallCond
	false, // YearBuilt // true?
	false, // YearRemodAdd // true
	true,  // RoofStyle
	true,  // RoofMatl
	true,  // Exterior1st
	true,  // Exterior2nd
	true,  // MasVnrType
	false, // MasVnrArea
	true,  // ExterQual
	true,  // ExterCond
	true,  // Foundation
	true,  // BsmtQual
	true,  // BsmtCond
	true,  // BsmtExposure
	true,  // BsmtFinType1
	false, // BsmtFinSF1
	true,  // BsmtFinType2
	false, // BsmtFinSF2
	false, // BsmtUnfSF
	false, // TotalBsmtSF
	true,  // Heating
	true,  // HeatingQC
	true,  // CentralAir
	true,  // Electrical
	false, // 1stFlrSF
	false, // 2ndFlrSF
	false, // LowQualFinSF
	false, // GrLivArea
	false, // BsmtFullBath
	false, // BsmtHalfBath
	false, // FullBath
	false, // HalfBath
	false, // BedroomAbvGr
	false, // KitchenAbvGr
	true,  // KitchenQual
	false, // TotRmsAbvGrd
	true,  // Functional
	false, // Fireplaces
	true,  // FireplaceQu
	true,  // GarageType
	false, // GarageYrBlt // true?
	true,  // GarageFinish
	false, // GarageCars
	false, // GarageArea
	true,  // GarageQual
	true,  // GarageCond
	true,  // PavedDrive
	false, // WoodDeckSF
	false, // OpenPorchSF
	false, // EnclosedPorch
	false, // 3SsnPorch
	false, // ScreenPorch
	false, // PoolArea
	true,  // PoolQC
	true,  // Fence
	true,  // MiscFeature
	false, // MiscVal
	false, // MoSold
	false, // YrSold // true?
	true,  // SaleType
	true,  // SaleCondition
	false, // SalePrice
}

var ignored = []string{
	"Id",
	// "MSSubClass",
	"MSZoning",
	// "LotFrontage",
	// "LotArea",
	"Street",
	"Alley",
	"LotShape",
	"LandContour",
	"Utilities",
	"LotConfig",
	"LandSlope",
	// "Neighborhood",
	"Condition1",
	"Condition2",
	"BldgType",
	// "HouseStyle",
	"OverallQual",
	"OverallCond",
	// "YearBuilt",
	// "YearRemodAdd",
	"RoofStyle",
	"RoofMatl",
	"Exterior1st",
	"Exterior2nd",
	"MasVnrType",
	// "MasVnrArea",
	"ExterQual",
	"ExterCond",
	// "Foundation",
	"BsmtQual",
	"BsmtCond",
	"BsmtExposure",
	"BsmtFinType1",
	// "BsmtFinSF1",
	"BsmtFinType2",
	// "BsmtFinSF2",
	// "BsmtUnfSF",
	// "TotalBsmtSF",
	// "Heating",
	"HeatingQC",
	// "CentralAir",
	"Electrical",
	// "1stFlrSF",
	// "2ndFlrSF",
	"LowQualFinSF",
	"GrLivArea",
	"BsmtFullBath",
	"BsmtHalfBath",
	// "FullBath",
	// "HalfBath",
	"BedroomAbvGr",
	"KitchenAbvGr",
	"KitchenQual",
	"TotRmsAbvGrd",
	"Functional",
	// "Fireplaces",
	"FireplaceQu",
	// "GarageType",
	"GarageYrBlt",
	"GarageFinish",
	"GarageCars",
	// "GarageArea",
	"GarageQual",
	"GarageCond",
	// "PavedDrive",
	// "WoodDeckSF",
	// "OpenPorchSF",
	"EnclosedPorch",
	"3SsnPorch",
	"ScreenPorch",
	// "PoolArea",
	"PoolQC",
	"Fence",
	"MiscFeature",
	"MiscVal",
	// "MoSold",
	// "YrSold",
	"SaleType",
	"SaleCondition",
	"SalePrice",
}

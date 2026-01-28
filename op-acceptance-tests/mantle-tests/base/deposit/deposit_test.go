package deposit

import "testing"

//	func TestDepositMNTByPortal_ZeroValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetMNT, msgValueZero)
//	}
//
//	func TestDepositMNTByPortal_WithoutMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetMNT, msgValueWithout)
//	}
//
//	func TestDepositMNTByPortal_WithMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetMNT, msgValueWith)
//	}
//
//	func TestDepositERC20ByPortal_ZeroValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetERC20, msgValueZero)
//	}
//
//	func TestDepositERC20ByPortal_WithoutMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetERC20, msgValueWithout)
//	}
//
//	func TestDepositERC20ByPortal_WithMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetERC20, msgValueWith)
//	}
//
//	func TestDepositETHByPortal_ZeroValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetETH, msgValueZero)
//	}
//
//	func TestDepositETHByPortal_WithoutMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetETH, msgValueWithout)
//	}
//
//	func TestDepositETHByPortal_WithMsgValue(gt *testing.T) {
//		runDepositCase(gt, depositByPortal, assetETH, msgValueWith)
//	}
//
//	func TestDepositMNTByBridge_ZeroValue(gt *testing.T) {
//		runDepositCase(gt, depositByBridge, assetMNT, msgValueZero)
//	}
func TestDepositMNTByBridge_WithoutMsgValue(gt *testing.T) {
	runDepositCase(gt, depositByBridge, assetMNT, msgValueWithout)
}

//
//func TestDepositMNTByBridge_WithMsgValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetMNT, msgValueWith)
//}
//
//func TestDepositERC20ByBridge_ZeroValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetERC20, msgValueZero)
//}
//
//func TestDepositERC20ByBridge_WithoutMsgValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetERC20, msgValueWithout)
//}
//
//func TestDepositERC20ByBridge_WithMsgValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetERC20, msgValueWith)
//}
//
//func TestDepositETHByBridge_ZeroValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetETH, msgValueZero)
//}
//
//func TestDepositETHByBridge_WithoutMsgValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetETH, msgValueWithout)
//}
//
//func TestDepositETHByBridge_WithMsgValue(gt *testing.T) {
//	runDepositCase(gt, depositByBridge, assetETH, msgValueWith)
//}

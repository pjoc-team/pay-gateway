package gateway

import pb "github.com/pjoc-team/pay-proto/go"

// SuccessResult success result
var SuccessResult = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS, Message: "SUCCESS", Describe: "SUCCESS"}

package gateway

import pb "github.com/pjoc-team/pay-proto/go"

var SUCCESS_RESULT = &pb.ReturnResult{Code: pb.ReturnResultCode_CODE_SUCCESS, Message: "SUCCESS", Describe: "SUCCESS"}

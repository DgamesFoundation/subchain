package peerex

import (
	pb "github.com/DgamesFoundation/subchain/fabric/protos"
	"golang.org/x/net/context"
)
//封装客户端的请求到链码请求
type rpcManager struct{
	ctx 	context.Context
	cancel  context.CancelFunc
}

type Rpc struct{
	
}

func (_ *Rpc) NewManager() *rpcManager{
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &rpcManager{ctx, cancel}
}

type RpcBuilder struct{
	ChaincodeName	 string
//	ChaincodeLang    string
	Function		 string
	
	Security		 *SecurityPolicy
	
	Conn			 ClientConn	
	ConnManager		 *rpcManager	
}

type SecurityPolicy struct{
	User			string
	Attributes		[]string
	Metadata		[]byte
	CustomIDGenAlg  string
}


var defaultSecPolicy = &SecurityPolicy{Attributes: []string{}}
// 客户端打包参数
func makeStringArgsToPb(funcname string, args []string) *pb.ChaincodeInput{
	
	input := &pb.ChaincodeInput{}
	//please remember the trick fabric used:
	//it push the "function name" as the first argument
	//in a rpc call
	var inarg [][]byte
	if len(funcname) == 0{
		input.Args = make([][]byte, len(args))	
		inarg = input.Args[:]
	}else{
		input.Args = make([][]byte, len(args) + 1)
		input.Args[0] = []byte(funcname)
		inarg = input.Args[1:] //引用Args[1:]
	}
	
	for i, arg := range args{
		inarg[i] = []byte(arg) //将其他参数赋给input.Args
	}
	
	return input
}


func (b *RpcBuilder) prepare(args []string) *pb.ChaincodeInvocationSpec{
	spec := &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG,	//always set it as golang
		ChaincodeID: &pb.ChaincodeID{Name: b.ChaincodeName},  // 链码名称
		CtorMsg : makeStringArgsToPb(b.Function, args),       // 组合之前设置Function自定义，args为编码后的args数组，参数1 地址，参数2 消息参数（转账接受者和币 或 注册对象） 参数3为 签名自签名
	}
	
	invocation := &pb.ChaincodeInvocationSpec{ChaincodeSpec: spec}
	
	if b.Security != nil{
		spec.SecureContext = b.Security.User                 //链码操作用户
		spec.Attributes = b.Security.Attributes              //链码权限数组
		spec.Metadata = b.Security.Metadata
		if len(b.Security.CustomIDGenAlg) != 0{              //ID生成算法
			invocation.IdGenerationAlg = b.Security.CustomIDGenAlg
		}
	}
	
	//final check attributes
	if spec.Attributes == nil{
		spec.Attributes = defaultSecPolicy.Attributes
	}
	
	return invocation	
}

func (b *RpcBuilder) context() context.Context{
		
	if b.ConnManager != nil{
		ctx, _ := context.WithCancel(b.ConnManager.ctx)
		return ctx
	}else{
		ctx, _ := context.WithCancel(context.Background())
		return ctx
	}
}
// 客户端 Fire Devops Invoke
func (b *RpcBuilder) Fire(args []string) (string, error){	
	//pb.NewDevopsClient(b.Conn.C) 获取rpc的client句柄 Invoke为pb的方法
	resp, err := pb.NewDevopsClient(b.Conn.C).Invoke(b.context(), b.prepare(args)) //链码调用  NewDevopsClient真正的grpc调用
	
	if err != nil{
		return "", err
	}
	
	return string(resp.Msg), nil
}
// 客户端 Query Devops Query
func (b *RpcBuilder) Query(args []string) ([]byte, error){	
	
	resp, err := pb.NewDevopsClient(b.Conn.C).Query(b.context(), b.prepare(args)) //通过封装的rpc查询结果
	
	if err != nil{
		return nil, err
	}
	
	return resp.Msg, nil
}

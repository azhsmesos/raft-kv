# raft-kv implement mit-8.624

## Raft 一致性算法将问题分成以下几个子问题
&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp; `leader选举(leader election)`  

&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp; `日志复制(log replication)`  

&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp; `安全性(safety)`  

&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp;&emsp; `成员关系变化(membership changes)`  


### 整体流程梳理
&emsp;&emsp;& ;1.默认以`follower`状态启动，租期号term = 1  
                      ;&emsp;&emsp;2.启动参数静态配置了集群的节点数量和节点地址 

&emsp;&emsp;&emsp;&emsp;3.`follower`和`candidate`节点不响应客户端请求  

&emsp;&emsp;&emsp;&emsp;4.等待`leader`心跳超时,进入`candidate`状态

&emsp;&emsp;&emsp;&emsp;5.向其他集群节点发起`RequestVote RPC`请求  

&emsp;&emsp;&emsp;&emsp;6.一个任期内，一个`raft`节点最多只能为一个候选人投票，先到先得  

&emsp;&emsp;&emsp;&emsp;7.如果`n/2 + 1`（包含自己那一票），则进入`leader`状态  

&emsp;&emsp;&emsp;&emsp;8.`leader`开始响应客户端请求，每条写请求都生成操作日志，并复制给其他节点  

&emsp;&emsp;&emsp;&emsp;9.大多数`follower`响应日志复制请求后，恢复ok  

&emsp;&emsp;&emsp;&emsp;10.`follower`提交操作日志到本地状态机  

&emsp;&emsp;&emsp;&emsp;11.`leader`租期结束，重新进入`candidate`状态    



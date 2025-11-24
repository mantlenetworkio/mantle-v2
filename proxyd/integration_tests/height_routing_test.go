package integration_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/ethereum-optimism/optimism/proxyd"
	"github.com/stretchr/testify/require"
)

// TestHeightBasedRouting 测试基于高度的路由功能
func TestHeightBasedRouting(t *testing.T) {
	// 初始化日志
	InitLogger()

	// 读取配置
	config := ReadConfig("height_routing")

	t.Run("route to primary for height >= cutoff", func(t *testing.T) {
		// 测试高度 >= 5000 的请求路由到 primary 后端
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		// Primary 返回高度 5000 的区块数据
		primaryRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x1388","hash":"0xprimary5000"}`)
		// Fallback 返回不同的数据（不应该被调用）
		fallbackRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x1388","hash":"0xfallback5000"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 请求高度 5000 的区块
		res, statusCode, err := client.SendRPC("eth_getBlockByNumber", []interface{}{"0x1388", false})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// 验证返回的是 primary 的数据
		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// Result 是字符串类型，需要再次解析
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "0xprimary5000", resultMap["hash"])

		// 验证只有 primary 被调用
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("route to fallback for height < cutoff", func(t *testing.T) {
		// 测试高度 < 5000 的请求路由到 fallback 后端
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		// Primary 返回数据（不应该被调用）
		primaryRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x1000","hash":"0xprimary4096"}`)
		// Fallback 返回历史数据
		fallbackRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x1000","hash":"0xfallback4096"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 请求高度 4096 的区块
		res, statusCode, err := client.SendRPC("eth_getBlockByNumber", []interface{}{"0x1000", false})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// 验证返回的是 fallback 的数据
		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// Result 是字符串类型，需要再次解析
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "0xfallback4096", resultMap["hash"])

		// 验证只有 fallback 被调用
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("fallback chain for hash queries", func(t *testing.T) {
		// 测试哈希查询：先查 primary，返回 null 则查 fallback
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		// Primary 返回 null（数据不存在）
		primaryRouter.SetFallbackRoute("eth_getBlockByHash", nil)
		// Fallback 返回历史数据
		fallbackRouter.SetFallbackRoute("eth_getBlockByHash", `{"number":"0x1000","hash":"0xabcd1234"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 通过哈希查询区块
		res, statusCode, err := client.SendRPC("eth_getBlockByHash", []interface{}{"0xabcd1234", false})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// 验证返回的是 fallback 的数据
		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// Result 是字符串类型，需要再次解析
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "0xabcd1234", resultMap["hash"])

		// 验证两个后端都被调用（fallback chain）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("route to primary for latest tag", func(t *testing.T) {
		// 测试 latest 标签路由到 primary
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x2000","hash":"0xlatest"}`)
		fallbackRouter.SetFallbackRoute("eth_getBlockByNumber", `{"number":"0x1000","hash":"0xold"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 请求 latest 区块
		res, statusCode, err := client.SendRPC("eth_getBlockByNumber", []interface{}{"latest", false})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		// 验证返回的是 primary 的数据
		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// Result 是字符串类型，需要再次解析
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "0xlatest", resultMap["hash"])

		// 验证只有 primary 被调用
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("route eth_getBalance with block parameter", func(t *testing.T) {
		// 测试 eth_getBalance 的路由（第二个参数是区块号）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		// 测试高度 < 5000 的余额查询
		fallbackRouter.SetFallbackRoute("eth_getBalance", "0x1234567890abcdef")
		primaryRouter.SetFallbackRoute("eth_getBalance", "0xfedcba0987654321")

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询高度 1000 的余额（< 5000，应该路由到 fallback）
		res, statusCode, err := client.SendRPC("eth_getBalance", []interface{}{
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			"0x3e8", // 1000
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.Equal(t, "0x1234567890abcdef", rpcRes.Result)

		// 验证只有 fallback 被调用
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("route range query entirely before cutoff to fallback", func(t *testing.T) {
		// 测试完全在 cutoff 之前的范围查询路由到 fallback
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x1000","logIndex":"0x0"}]`)
		primaryRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [1000, 2000]，完全在 5000 之前
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x3e8", // 1000
				"toBlock":   "0x7d0", // 2000
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("route range query entirely after cutoff to primary", func(t *testing.T) {
		// 测试完全在 cutoff 之后的范围查询路由到 primary
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x2000","logIndex":"0x0"}]`)
		fallbackRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [6000, 8000]，完全在 5000 之后
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x1770", // 6000
				"toBlock":   "0x1f40", // 8000
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 primary 被调用
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("route range query spanning cutoff to fallback", func(t *testing.T) {
		// 测试跨越 cutoff 的范围查询路由到 fallback（因为它有全量数据）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x1000","logIndex":"0x0"},{"blockNumber":"0x2000","logIndex":"0x0"}]`)
		primaryRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [4000, 6000]，跨越 cutoff 5000
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0xfa0",  // 4000
				"toBlock":   "0x1770", // 6000
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用（跨越 cutoff 的查询使用有全量数据的 fallback）
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("batch requests with mixed heights", func(t *testing.T) {
		// 测试混合高度的批量请求
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		// Primary 和 Fallback 都需要设置所有可能的响应
		primaryRouter.SetRoute("eth_getBlockByNumber", "1", `{"number":"0x1388","hash":"0xprimary5000"}`)
		primaryRouter.SetRoute("eth_getBlockByNumber", "2", `{"number":"0x1000","hash":"0xprimary4096"}`)

		fallbackRouter.SetRoute("eth_getBlockByNumber", "1", `{"number":"0x1388","hash":"0xfallback5000"}`)
		fallbackRouter.SetRoute("eth_getBlockByNumber", "2", `{"number":"0x1000","hash":"0xfallback4096"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 批量请求：一个 >= cutoff，一个 < cutoff
		reqs := []*proxyd.RPCReq{
			NewRPCReq("1", "eth_getBlockByNumber", []interface{}{"0x1388", false}), // 5000
			NewRPCReq("2", "eth_getBlockByNumber", []interface{}{"0x1000", false}), // 4096
		}

		res, statusCode, err := client.SendBatchRPC(reqs...)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var batchRes []proxyd.RPCRes
		err = json.Unmarshal(res, &batchRes)
		require.NoError(t, err)
		require.Equal(t, 2, len(batchRes))

		// 验证第一个请求（5000）的响应
		if batchRes[0].Result != nil {
			resultStr1, ok := batchRes[0].Result.(string)
			if ok {
				var result1 map[string]interface{}
				json.Unmarshal([]byte(resultStr1), &result1)
				fmt.Printf("Result 1 hash: %v\n", result1["hash"])
			}
		}

		// 验证第二个请求（4096）的响应
		if batchRes[1].Result != nil {
			resultStr2, ok := batchRes[1].Result.(string)
			if ok {
				var result2 map[string]interface{}
				json.Unmarshal([]byte(resultStr2), &result2)
				fmt.Printf("Result 2 hash: %v\n", result2["hash"])
			}
		}

		// 验证两个后端都被调用了（因为批量请求混合了不同范围的高度）
		fmt.Printf("Primary requests: %d, Fallback requests: %d\n",
			len(primaryBackend.Requests()), len(fallbackBackend.Requests()))
	})

	t.Run("getLogs with latest tag spanning cutoff", func(t *testing.T) {
		// 测试使用 latest 标签的范围查询（跨越 cutoff）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x1000","logIndex":"0x0"},{"blockNumber":"0x2000","logIndex":"0x0"}]`)
		primaryRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [4096, latest]，跨越 cutoff 5000
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x1000", // 4096
				"toBlock":   "latest", // MaxUint64
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用（跨越 cutoff 使用 fallback）
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("getLogs without block range defaults", func(t *testing.T) {
		// 测试省略 fromBlock 和 toBlock 的情况（默认 earliest to latest）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x100","logIndex":"0x0"}]`)
		primaryRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 不指定 fromBlock 和 toBlock，默认查询全范围
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
				"topics":  []interface{}{"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"},
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用（默认 0 to MaxUint64 跨越 cutoff）
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("getLogs with earliest tag", func(t *testing.T) {
		// 测试使用 earliest 标签的范围查询
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x500","logIndex":"0x0"}]`)
		primaryRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [earliest, 4096]，完全在 cutoff 之前
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "earliest",
				"toBlock":   "0x1000", // 4096
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用（完全在 cutoff 之前）
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("getLogs with finalized tag", func(t *testing.T) {
		// 测试使用 finalized 标签的范围查询（完全在 cutoff 之后）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x2000","logIndex":"0x0"}]`)
		fallbackRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [8192, finalized]，完全在 cutoff 之后
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x2000",    // 8192
				"toBlock":   "finalized", // MaxUint64
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 primary 被调用（完全在 cutoff 之后）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("eth_newFilter with range query before cutoff", func(t *testing.T) {
		// 测试 eth_newFilter 方法（使用与 getLogs 相同的过滤器结构）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("eth_newFilter", "0x1") // 返回 filter ID
		primaryRouter.SetFallbackRoute("eth_newFilter", "0x2")

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 创建过滤器：范围 [1000, 2000]，完全在 cutoff 之前
		res, statusCode, err := client.SendRPC("eth_newFilter", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x3e8", // 1000
				"toBlock":   "0x7d0", // 2000
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 fallback 被调用
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("getLogs boundary case at exactly cutoff height", func(t *testing.T) {
		// 测试边界情况：正好在 cutoff 高度上
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x1388","logIndex":"0x0"}]`)
		fallbackRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [5000, 5000]，正好在 cutoff 上
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x1388", // 5000
				"toBlock":   "0x1388", // 5000
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 primary 被调用（高度 >= cutoff 路由到 primary）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("getLogs with safe tag after cutoff", func(t *testing.T) {
		// 测试使用 safe 标签的范围查询
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getLogs", `[{"blockNumber":"0x1500","logIndex":"0x0"}]`)
		fallbackRouter.SetFallbackRoute("eth_getLogs", `[]`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 查询范围 [6000, safe]，完全在 cutoff 之后
		res, statusCode, err := client.SendRPC("eth_getLogs", []interface{}{
			map[string]interface{}{
				"fromBlock": "0x1770", // 6000
				"toBlock":   "safe",   // MaxUint64
				"address":   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证只有 primary 被调用（完全在 cutoff 之后）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("custom method with parameter at index 1", func(t *testing.T) {
		// 测试自定义方法，区块参数在第二个位置（索引1）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		fallbackRouter.SetFallbackRoute("custom_getDataAtBlock", `{"data":"historical_data"}`)
		primaryRouter.SetFallbackRoute("custom_getDataAtBlock", `{"data":"recent_data"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 自定义方法: custom_getDataAtBlock(contract_address, block_height)
		// 参数索引1（第二个参数）是区块号，值为 0x1000 (4096 < 5000)
		res, statusCode, err := client.SendRPC("custom_getDataAtBlock", []interface{}{
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", // 合约地址
			"0x1000", // 4096 - 区块高度（在索引1）
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// 验证返回的是 fallback 的数据（因为高度 < cutoff）
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "historical_data", resultMap["data"])

		// 验证只有 fallback 被调用
		require.Equal(t, 0, len(primaryBackend.Requests()))
		require.Equal(t, 1, len(fallbackBackend.Requests()))
	})

	t.Run("custom method with parameter at index 0", func(t *testing.T) {
		// 测试自定义方法，区块参数在第一个位置（索引0）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("custom_queryByHeight", `{"result":"primary_query"}`)
		fallbackRouter.SetFallbackRoute("custom_queryByHeight", `{"result":"fallback_query"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 自定义方法: custom_queryByHeight(block_height, query_data)
		// 参数索引0（第一个参数）是区块号，值为 0x1770 (6000 >= 5000)
		res, statusCode, err := client.SendRPC("custom_queryByHeight", []interface{}{
			"0x1770",      // 6000 - 区块高度（在索引0）
			"0xquerydata", // 查询数据
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.NotNil(t, rpcRes.Result)

		// 验证返回的是 primary 的数据（因为高度 >= cutoff）
		resultStr, ok := rpcRes.Result.(string)
		require.True(t, ok, "Result should be a string")

		var resultMap map[string]interface{}
		err = json.Unmarshal([]byte(resultStr), &resultMap)
		require.NoError(t, err)
		require.Equal(t, "primary_query", resultMap["result"])

		// 验证只有 primary 被调用
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("custom method with parameter at index 2", func(t *testing.T) {
		// 测试自定义方法，区块参数在第三个位置（索引2）
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("eth_getBalance_custom", "0xprimarybalance")
		fallbackRouter.SetFallbackRoute("eth_getBalance_custom", "0xfallbackbalance")

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 自定义方法: eth_getBalance_custom(address, extra_param, block_height)
		// 参数索引2（第三个参数）是区块号，值为 0x2000 (8192 >= 5000)
		res, statusCode, err := client.SendRPC("eth_getBalance_custom", []interface{}{
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", // 地址
			"extra_parameter", // 额外参数
			"0x2000",          // 8192 - 区块高度（在索引2）
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)
		require.Equal(t, "0xprimarybalance", rpcRes.Result)

		// 验证只有 primary 被调用（因为高度 >= cutoff）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("custom method with boundary height at cutoff", func(t *testing.T) {
		// 测试自定义方法在 cutoff 边界的行为
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("custom_queryByHeight", `{"boundary":"at_cutoff"}`)
		fallbackRouter.SetFallbackRoute("custom_queryByHeight", `{"boundary":"before_cutoff"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 测试正好在 cutoff 高度 5000
		res, statusCode, err := client.SendRPC("custom_queryByHeight", []interface{}{
			"0x1388", // 5000 - 正好等于 cutoff
			"0xdata",
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证路由到 primary（高度 >= cutoff）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})

	t.Run("custom method with latest tag", func(t *testing.T) {
		// 测试自定义方法使用 latest 标签
		primaryRouter := NewBatchRPCResponseRouter()
		fallbackRouter := NewBatchRPCResponseRouter()

		primaryRouter.SetFallbackRoute("custom_getDataAtBlock", `{"data":"latest_data"}`)
		fallbackRouter.SetFallbackRoute("custom_getDataAtBlock", `{"data":"old_data"}`)

		primaryBackend := NewMockBackend(primaryRouter)
		fallbackBackend := NewMockBackend(fallbackRouter)
		defer primaryBackend.Close()
		defer fallbackBackend.Close()

		require.NoError(t, os.Setenv("PRIMARY_BACKEND_RPC_URL", primaryBackend.URL()))
		require.NoError(t, os.Setenv("FALLBACK_BACKEND_RPC_URL", fallbackBackend.URL()))

		client := NewProxydClient("http://127.0.0.1:8545")
		_, shutdown, err := proxyd.Start(config)
		require.NoError(t, err)
		defer shutdown()

		// 使用 latest 标签
		res, statusCode, err := client.SendRPC("custom_getDataAtBlock", []interface{}{
			"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			"latest", // latest 标签应该路由到 primary
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		var rpcRes proxyd.RPCRes
		err = json.Unmarshal(res, &rpcRes)
		require.NoError(t, err)
		require.Nil(t, rpcRes.Error)

		// 验证路由到 primary（latest 视为最新数据）
		require.Equal(t, 1, len(primaryBackend.Requests()))
		require.Equal(t, 0, len(fallbackBackend.Requests()))
	})
}

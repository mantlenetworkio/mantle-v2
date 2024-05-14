package deploy

import "reflect"

type DisperserVars struct {
	DL_DISPERSER_HOSTNAME string

	DL_DISPERSER_GRPC_PORT string

	DL_DISPERSER_ENABLE_METRICS string

	DL_DISPERSER_METRICS_PORT string

	DL_DISPERSER_TIMEOUT string

	DL_DISPERSER_POLLING_RETRY string

	DL_DISPERSER_DB_PATH string

	DL_DISPERSER_GRAPH_PROVIDER string

	DL_DISPERSER_DLSM_ADDRESS string

	DL_DISPERSER_USE_CACHE string

	DL_DISPERSER_CHAIN_RPC string

	DL_DISPERSER_PRIVATE_KEY string

	DL_DISPERSER_CHAIN_ID string

	DL_DISPERSER_STD_LOG_LEVEL string

	DL_DISPERSER_FILE_LOG_LEVEL string

	DL_DISPERSER_LOG_PATH string

	DL_DISPERSER_G1_PATH string

	DL_DISPERSER_G2_PATH string

	DL_DISPERSER_CACHE_PATH string

	DL_DISPERSER_SRS_ORDER string

	DL_DISPERSER_NUM_WORKERS string

	DL_DISPERSER_VERBOSE string
}

func (vars DisperserVars) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}

type OperatorVars struct {
	DL_NODE_HOSTNAME string

	DL_NODE_GRPC_PORT string

	DL_NODE_ENABLE_METRICS string

	DL_NODE_METRICS_PORT string

	DL_NODE_TIMEOUT string

	DL_NODE_DB_PATH string

	DL_NODE_GRAPH_PROVIDER string

	DL_NODE_PRIVATE_BLS string

	DL_NODE_DLSM_ADDRESS string

	DL_NODE_CHALLENGE_ORDER string

	DL_NODE_CHAIN_RPC string

	DL_NODE_PRIVATE_KEY string

	DL_NODE_CHAIN_ID string

	DL_NODE_STD_LOG_LEVEL string

	DL_NODE_FILE_LOG_LEVEL string

	DL_NODE_LOG_PATH string

	DL_NODE_G1_PATH string

	DL_NODE_G2_PATH string

	DL_NODE_CACHE_PATH string

	DL_NODE_SRS_ORDER string

	DL_NODE_NUM_WORKERS string

	DL_NODE_VERBOSE string
}

func (vars OperatorVars) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}

type RetrieverVars struct {
	DL_RETRIEVER_HOSTNAME string

	DL_RETRIEVER_GRPC_PORT string

	DL_RETRIEVER_TIMEOUT string

	DL_RETRIEVER_GRAPH_PROVIDER string

	DL_RETRIEVER_DLSM_ADDRESS string

	DL_RETRIEVER_CHAIN_RPC string

	DL_RETRIEVER_PRIVATE_KEY string

	DL_RETRIEVER_CHAIN_ID string

	DL_RETRIEVER_STD_LOG_LEVEL string

	DL_RETRIEVER_FILE_LOG_LEVEL string

	DL_RETRIEVER_LOG_PATH string

	DL_RETRIEVER_G1_PATH string

	DL_RETRIEVER_G2_PATH string

	DL_RETRIEVER_CACHE_PATH string

	DL_RETRIEVER_SRS_ORDER string

	DL_RETRIEVER_NUM_WORKERS string

	DL_RETRIEVER_VERBOSE string
}

func (vars RetrieverVars) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}

type RollupSequencerVars struct {
	SEQUENCER_DISPERSER string

	SEQUENCER_GRPC_PORT string

	SEQUENCER_GRAPH_PROVIDER string

	SEQUENCER_ROLLUP_ADDRESS string

	SEQUENCER_DURATION string

	SEQUENCER_TIMEOUT string

	SEQUENCER_CHAIN_RPC string

	SEQUENCER_PRIVATE_KEY string

	SEQUENCER_CHAIN_ID string

	SEQUENCER_STD_LOG_LEVEL string

	SEQUENCER_FILE_LOG_LEVEL string

	SEQUENCER_LOG_PATH string
}

func (vars RollupSequencerVars) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}

type RollupChallengerVars struct {
	CHALLENGER_RETRIEVER string

	CHALLENGER_GRAPH_PROVIDER string

	CHALLENGER_ROLLUP_ADDRESS string

	CHALLENGER_G1_PATH string

	CHALLENGER_G2_PATH string

	CHALLENGER_CACHE_PATH string

	CHALLENGER_SRS_ORDER string

	CHALLENGER_KZG_NUM_WORKERS string

	CHALLENGER_TIMEOUT string

	CHALLENGER_CHAIN_RPC string

	CHALLENGER_PRIVATE_KEY string

	CHALLENGER_CHAIN_ID string

	CHALLENGER_STD_LOG_LEVEL string

	CHALLENGER_FILE_LOG_LEVEL string

	CHALLENGER_LOG_PATH string
}

func (vars RollupChallengerVars) getEnvMap() map[string]string {
	v := reflect.ValueOf(vars)
	envMap := make(map[string]string)
	for i := 0; i < v.NumField(); i++ {
		envMap[v.Type().Field(i).Name] = v.Field(i).String()
	}
	return envMap
}

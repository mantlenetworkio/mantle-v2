import { BigInt, Address,  Bytes, ethereum, crypto, ByteArray } from "@graphprotocol/graph-ts";
import { log } from '@graphprotocol/graph-ts'
import { BLSRegistry, Deregistration, Registration, SocketUpdate, StakeUpdate } from "../generated/BLSRegistry/BLSRegistry";
import { OperatorPublicKeys, OperatorIndex, OperatorStake, Operator, TotalStake, TotalOperator, Total } from '../generated/schema'

const MAX_UINT32 = BigInt.fromU32(2**32-1)

export function handleRegistration(event: Registration): void {
    log.debug('here', [])
    let operatorId = event.params.operator.toHex()
    let op = new Operator(operatorId)
    let pubkeys = OperatorPublicKeys.load(operatorId)
    if (!pubkeys) {
        throw new Error("handleRegistration: pubkeys not found")
    }
    op.pubkeys = pubkeys.id
    let registry = BLSRegistry.bind(event.address)
    op.fromBlockNumber  = event.block.number
    op.toBlockNumber = MAX_UINT32
    op.socket = event.params.socket
    op.indexHistoryIds = []
    op.stakeHistoryIds = []
    op.status = 0
    op.save()

    // Create opStakesHistory
    //id for index and stakes are (registrant || num) where num is the number of stakes/indecies for that operator minus 1
    let opStakeAndIndexId = operatorId + "0"

    let opStake = new OperatorStake(opStakeAndIndexId)
    opStake.operator = operatorId

    //next update block is zero because another update has not happened
    opStake.toBlockNumber = MAX_UINT32

    // let registry = BLSRegistry.bind(event.address)
    //since this is registration, get the first stake
    let stake = registry.pubkeyHashToStakeHistory(event.params.pkHash, BigInt.zero())
    opStake.mantleFirstStake = stake.getFirstQuorumStake()
    opStake.mantleSencodStake = stake.getSecondQuorumStake()

    // Create opIndexHistory
    //since this is registration, get the first index
    let index = registry.pubkeyHashToIndexHistory(event.params.pkHash, BigInt.zero())
    let opIndex = new OperatorIndex(opStakeAndIndexId)

    opIndex.operator = operatorId
    opIndex.toBlockNumber = MAX_UINT32
    opIndex.index = index.getIndex()

    // update
    let opIds = Operator.load(operatorId)
    if (!opIds) {
      throw new Error("handleDeregistration: Operator not found")
    }
    let opStakeHistoryIds = op.stakeHistoryIds
    opStakeHistoryIds.push(opStakeAndIndexId)
    opIds.stakeHistoryIds = opStakeHistoryIds
    let opIndexHistoryIds = op.indexHistoryIds
    opIndexHistoryIds.push(opStakeAndIndexId)
    opIds.indexHistoryIds = opIndexHistoryIds
    opIds.save()

    opStake.save()
    opIndex.save()

    let totals = getTotal()

    // Update TotalStake

    let stakeHistory = totals.stakeHistory

    let prevTotalStake = TotalStake.load((stakeHistory.length - 1).toString())
    if (!prevTotalStake) {
        throw new Error("handleRegistration: prevTotalStake not found")
    }
    prevTotalStake.toBlockNumber = event.block.number
    prevTotalStake.save()

    let totalStake = new TotalStake(stakeHistory.length.toString())
    //read stakes from event
    totalStake.mantleFirstStake = prevTotalStake.mantleFirstStake.plus(opStake.mantleFirstStake)
    totalStake.mantleSencodStake = prevTotalStake.mantleSencodStake.plus(opStake.mantleSencodStake)
    totalStake.toBlockNumber = MAX_UINT32
    totalStake.index = prevTotalStake.index + 1
    totalStake.save()

    stakeHistory.push(totalStake.id)
    totals.stakeHistory = stakeHistory

    // Update TotalOperators
    let indexHistory = totals.indexHistory
    let prevTotalOperator = TotalOperator.load((indexHistory.length - 1).toString())
    if (!prevTotalOperator) {

        let initTotalOperator = new TotalOperator("0")
        initTotalOperator.save()

        throw new Error("handleRegistration: prevTotalStake not found")
    }
    prevTotalOperator.toBlockNumber = event.block.number
    prevTotalOperator.save()

    let totalOperators = new TotalOperator(indexHistory.length.toString())
    //read stakes from event
    totalOperators.count = prevTotalOperator.count.plus(BigInt.fromU32(1))

    let apk:BigInt[];
    apk = []
    // let apk = []
    apk.push(registry.apk().getX())
    apk.push(registry.apk().getY())

    totalOperators.aggPubKey = apk
    totalOperators.aggPubKeyHash = event.params.apkHash
    totalOperators.index = prevTotalOperator.index + 1
    totalOperators.toBlockNumber = MAX_UINT32
    totalOperators.save()

    indexHistory.push(totalOperators.id)
    totals.indexHistory = indexHistory

    totals.save()
}

function getTotal(): Total {

    let totals = Total.load("0")
    if (!totals){
        totals = new Total("0")

        let totalStake = new TotalStake("0")
        totalStake.index = 0
        totalStake.mantleFirstStake = BigInt.zero()
        totalStake.mantleSencodStake = BigInt.zero()
        totalStake.toBlockNumber = BigInt.zero()
        totalStake.save()
        totals.stakeHistory = [totalStake.id]

        let totalOperators = new TotalOperator("0")
        totalOperators.index = 0
        totalOperators.aggPubKey = []
        totalOperators.aggPubKeyHash = Bytes.empty()
        totalOperators.count = BigInt.zero()
        totalOperators.toBlockNumber = MAX_UINT32
        totalOperators.save()
        totals.indexHistory = [totalOperators.id]

        totals.save()
    }

    return totals
}

export function handleSocketUpdate(event: SocketUpdate): void {
    let op = Operator.load(event.params.operator.toHex())
    if (!op) {
        throw new Error("handleSocketUpdate: Operator not found")
    }
    op.socket = event.params.socket
    op.save()
}

export function handleStakeUpdate(event: StakeUpdate): void {
    let op = Operator.load(event.params.operator.toHex())
    if (!op) {
        throw new Error("handleStakeUpdate: Operator not found")
    }

    //create opStake object
    let operatorId = event.params.operator.toHex()
    let opStake = new OperatorStake(operatorId + op.stakeHistoryIds.length.toString())
    //read stakes from event
    opStake.operator = operatorId
    opStake.mantleFirstStake = event.params.firstQuorumStake
    opStake.mantleSencodStake = event.params.secondQuorumStake
    //default block number for new stake
    opStake.toBlockNumber = MAX_UINT32

    let opStakeIds = op.stakeHistoryIds;
    opStakeIds.push(operatorId + op.stakeHistoryIds.length.toString())
    op.stakeHistoryIds = opStakeIds

    op.save()
    opStake.save()

    let prevOpStake = OperatorStake.load(operatorId + (op.stakeHistoryIds.length - 1).toString())
    if (!prevOpStake) {
        throw new Error("handleStakeUpdate: prevOpStake not found")
    }
    //set update from the last operator stake
    prevOpStake.toBlockNumber = event.block.number
    prevOpStake.save()

    //save new stake and add to operator
    opStake.save()
}

export function handleDeregistration(event: Deregistration): void {
    let operatorId = event.params.operator.toHex()

    let op = Operator.load(operatorId)
    if (!op) {
        throw new Error("handleDeregistration: Operator not found")
    }

    op.toBlockNumber = event.block.number
    op.status = 1

    //create opStake object
    let opStake = new OperatorStake(operatorId + op.stakeHistoryIds.length.toString())
    opStake.operator = operatorId
    //read stakes from event
    opStake.mantleFirstStake = BigInt.zero()
    opStake.mantleSencodStake = BigInt.zero()
    //default block number for new stake
    opStake.toBlockNumber = MAX_UINT32
    let opStakeIds = op.stakeHistoryIds;
    opStakeIds.push(operatorId + op.stakeHistoryIds.length.toString())
    op.stakeHistoryIds = opStakeIds

    opStake.save()

    op.save()

    let prevOpStake = OperatorStake.load(operatorId + (op.stakeHistoryIds.length - 1).toString())
    if (!prevOpStake) {
        throw new Error("handleStakeUpdate: prevOpStake not found")
    }
    //set update from the last operator stake
    prevOpStake.toBlockNumber = event.block.number

    //save new stake and add to operator
    opStake.save()
    // op.stakeHistory.push(opStake.id)

    //update the index
    let prevOpIndex = OperatorIndex.load(operatorId + (op.indexHistoryIds.length - 1).toString())
    if (!prevOpIndex) {
        throw new Error("handleStakeUpdate: prevOpIndex not found")
    }

    //set update from the last operator index
    prevOpIndex.toBlockNumber = event.block.number
    prevOpIndex.save()

    let totals = getTotal()

    // Update TotalStake
    let stakeHistory = totals.stakeHistory
    let prevTotalStake = TotalStake.load((stakeHistory.length - 1).toString())
    if (!prevTotalStake) {
      throw new Error("handleRegistration: prevTotalStake not found")
    }
    prevTotalStake.toBlockNumber = event.block.number
    prevTotalStake.save()

    let totalStake = new TotalStake(stakeHistory.length.toString())
    //read stakes from event
    totalStake.mantleFirstStake = prevTotalStake.mantleFirstStake.minus(prevOpStake.mantleFirstStake)
    totalStake.mantleSencodStake = prevTotalStake.mantleSencodStake.minus(prevOpStake.mantleSencodStake)
    totalStake.toBlockNumber = MAX_UINT32
    totalStake.index = prevTotalStake.index - 1
    totalStake.save()
    stakeHistory.push(totalStake.id)
    totals.stakeHistory = stakeHistory

    // Update TotalOperators
    let indexHistory = totals.indexHistory
    let prevTotalOperator = TotalOperator.load((indexHistory.length - 1).toString())
    if (!prevTotalOperator) {
      throw new Error("handleRegistration: prevTotalStake not found")
    }
    prevTotalOperator.toBlockNumber = event.block.number
    prevTotalOperator.save()

    let totalOperators = new TotalOperator(indexHistory.length.toString())
    //read stakes from event
    totalOperators.count = prevTotalOperator.count.minus(BigInt.fromU32(1))
    let registry = BLSRegistry.bind(event.address)
    let apk:BigInt[];
    apk = []
    apk.push(registry.apk().getX())
    apk.push(registry.apk().getY())
    totalOperators.aggPubKey = apk
    totalOperators.aggPubKeyHash = event.params.apkHash
    totalOperators.index = prevTotalOperator.index - 1
    totalOperators.toBlockNumber = MAX_UINT32
    totalOperators.save()

    indexHistory.push(totalOperators.id)
    totals.indexHistory = indexHistory

    totals.save()
    prevOpStake.save()

    //if there is someone who was swapped, update their index
    if (event.params.swapped != Address.fromHexString("0x0000000000000000000000000000000000000000")) {
        let registry = BLSRegistry.bind(event.address)
        let pubkeyHash = registry.registry(event.params.swapped).getPubkeyHash()

        let swapped = Operator.load(event.params.swapped.toHex())
        if (!swapped) {
            throw new Error("handleDeregistration: swapped not found")
        }

        //create opIndex object
        let swappedId = event.params.swapped.toHex()
        let swappedIndex = new OperatorIndex(swappedId + swapped.indexHistoryIds.length.toString())
        //get corresponding OperatorIndex from chain and store
        swappedIndex.operator = swappedId
        let index = registry.pubkeyHashToIndexHistory(pubkeyHash, BigInt.fromI32(swapped.indexHistoryIds.length))
        swappedIndex.index = index.getIndex()
        swappedIndex.toBlockNumber = MAX_UINT32

        swappedIndex.save()

        //get prev index and update the toBlockNumber
        let prevSwappedIndex = OperatorIndex.load(event.params.swapped.toHex() + (swapped.indexHistoryIds.length - 1).toString())
        if (!prevSwappedIndex) {
            throw new Error("handleDeregistration: prevSwappedIndex not found")
        }
        prevSwappedIndex.toBlockNumber = event.block.number
        prevSwappedIndex.save()

        // add swapped index to history and save
        // swapped.indexHistory.push(swappedIndex.id)
        swapped.save()
    }
}

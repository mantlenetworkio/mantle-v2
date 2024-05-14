import { BigInt, Address, ethereum, Bytes, ByteArray } from "@graphprotocol/graph-ts";
import { ConfirmDataStore, DataLayrServiceManager, InitDataStore, SignatoryRecord, InitDataStoreCall } from '../generated/DataLayrServiceManager/DataLayrServiceManager'
import { DataStore } from '../generated/schema'
import { crypto , log } from '@graphprotocol/graph-ts'

export function handleInitDataStore(event: InitDataStore): void {

    log.info("handleInitDataStore",[])

    let searchData = event.params.searchData
    let metadata = searchData.metadata
    
    let ds = new DataStore(metadata.globalDataStoreId.toString()) //event.params.headerHash.toHex()
    //keccak256(abi.encodePacked(dataStoreIdToConfirm, searchData.metadata.headerHash, searchData.duration, searchData.timestamp, searchData.index))
    //4, 32, 1, 32, 4
    ds.storeNumber = metadata.globalDataStoreId
    ds.durationDataStoreId = metadata.durationDataStoreId
    ds.index = searchData.index
    ds.dataCommitment = metadata.headerHash
    ds.referenceBlockNumber = metadata.referenceBlockNumber
    ds.fee = metadata.fee
    ds.initTxHash = event.transaction.hash
	ds.initBlockNumber = event.block.number
    ds.confirmer = metadata.confirmer.toHexString()

    ds.headerHash = metadata.headerHash
    ds.header = event.params.header

    // Decode Header
    let read = (data:Bytes, len: i32):Bytes=> (Bytes.fromHexString(data.toHex().slice(2).slice(0,len*2)))
    let readRev = (data:Bytes, len: i32):Bytes=> (Bytes.fromUint8Array(Bytes.fromHexString(data.toHex().slice(2).slice(0,len*2)).reverse()))
    let trim = (data:Bytes, len: i32):Bytes => (Bytes.fromHexString(data.toHex().slice(2).slice(len*2)))

    let header = event.params.header

    ds.kzgCommit = read(header,64)
    header = trim(header,64)

    ds.degree = BigInt.fromU32(readRev(header,4).toU32())
    header = trim(header,4)
    
    ds.numSys = BigInt.fromU32(readRev(header,4).toU32())
    header = trim(header,4)

    ds.numPar = BigInt.fromU32(readRev(header,4).toU32())
    header = trim(header,4)

    ds.origDataSize = BigInt.fromU32(readRev(header,4).toU32())
    header = trim(header,4)

    ds.disperser = read(header,20)
    header = trim(header,20)

    ds.lowDegreeProof = read(header,64)
    
    ds.initTime = event.block.timestamp
    ds.duration = searchData.duration

    let dlsm = DataLayrServiceManager.bind(event.address)

    ds.storePeriodLength = (BigInt.fromU32(searchData.duration)).times(dlsm.DURATION_SCALE())
    // add an extra 32 blocks for reorg tolerance
    ds.expireTime = event.block.timestamp.plus(dlsm.confirmDataStoreTimeout()).plus(BigInt.fromU32(12*32))
    
    ds.initGasUsed = BigInt.zero()
    if(event.receipt) {
        if (event.receipt!.gasUsed !== null){
            ds.initGasUsed = event.receipt!.gasUsed
        }else{
            log.info("gasUsed null",[])
        }
    }

    // let durationBigInt = event.params.storePeriodLength.div(DataLayrServiceManager.bind(event.address).DURATION_SCALE())
    // ds.duration = durationBigInt.toI32()

    ds.msgHash = Bytes.fromByteArray(crypto.keccak256(ByteArray.fromHexString(bigIntToHexPad(metadata.globalDataStoreId, 8) + bytesToHexPad(metadata.headerHash, 64) + bigIntToHexPad(BigInt.fromU32(searchData.duration), 2) + bigIntToHexPad(event.block.timestamp, 64) + bigIntToHexPad(searchData.index, 8))))

    // filled in by confirm 
    ds.confirmed = false
    
    ds.lastUpdated = event.block.timestamp
    ds.save()
}

export function handleSignatoryRecord(event: SignatoryRecord): void {
    let ds = DataStore.load(event.params.taskNumber.toString())
    if (!ds) {
        throw new Error(`No metadata found corresponding to confirm`);
    }

    ds.confirmed = true
    ds.ethSigned = event.params.signedStakeFirstQuorum
    ds.eigenSigned = event.params.signedStakeSecondQuorum
    ds.confirmTxHash = event.transaction.hash
    ds.nonSignerPubKeyHashes = event.params.pubkeyHashes
    ds.signatoryRecord = Bytes.fromByteArray(crypto.keccak256(ByteArray.fromHexString(bigIntToHexPad(event.params.taskNumber, 8) + bytesArrayToHexPad(event.params.pubkeyHashes, 64) + bigIntToHexPad(event.params.signedStakeFirstQuorum, 64) + bigIntToHexPad(event.params.signedStakeSecondQuorum, 64))))

    ds.confirmGasUsed = BigInt.zero()
    if(event.receipt) {
        if (event.receipt!.gasUsed !== null){
            ds.confirmGasUsed = event.receipt!.gasUsed
        }else{
            log.info("gasUsed null",[])
        }
    }

    ds.expireTime = event.block.timestamp.plus(ds.storePeriodLength)
    ds.lastUpdated = event.block.timestamp
    ds.save()
}

// f(255, 5) = "000FF"
function bigIntToHexPad(x: BigInt, targetLength: i32): string {
    return x.toHex().substr(2).padStart(targetLength, "0")
}

// f(<FF, FF>, 5) = "000FF"
function bytesToHexPad(x: Bytes, targetLength: i32): string {
    return x.toHex().substr(2).padStart(targetLength, "0")
}


function bytesArrayToHexPad(xs: Bytes[], targetLength: i32): string {
    let hex = ""
    for (let i=0; i<xs.length; i++){
        hex += bytesToHexPad(xs[i],targetLength)
    }
    return hex
}
import {
  getERC20Balance,
  getERC20Holders,
  getERC721Holders,
  getERC721Tokens,
  getERC721TokensForAccount,
  getEvents,
  getHolderForERC721Token,
  getInternalToTxs,
  getReportData,
  getToTxs
} from './client/fetcher'
import { TransactionTable } from './components/table/Transactions'
import { EventTable } from './components/table/Events'
import { ReportTable } from './components/table/Report'
import { TokenBalanceTable, TokenHolderTable, TokenTable } from './components/table/Tokens'

export const Reports = {
  GenerateReport: {
    label: 'Full Report',
    value: 'GenerateReport',
    fields: {
      startBlock: 'optional',
      endBlock: 'optional',
    },
    View: ReportTable,
    getItems: (rpcEndpoint, params, options) => getReportData(rpcEndpoint, params.address, params.startBlockNumber, params.endBlockNumber, options),
  },
  ERC20TokenBalance: {
    label: 'ERC20 Token Balance',
    value: 'ERC20TokenBalance',
    fields: {
      account: 'required',
      startBlock: 'optional',
      endBlock: 'optional',
    },
    View: TokenBalanceTable,
    getItems: (rpcEndpoint, params, options) => getERC20Balance(rpcEndpoint, params.address, params.account, params.startBlockNumber, params.endBlockNumber, options),
  },
  ERC20TokenHolders: {
    label: 'ERC20 Token Holders',
    value: 'ERC20TokenHolders',
    fields: {
      block: 'optional',
    },
    View: TokenHolderTable,
    getItems: (rpcEndpoint, params, options) => getERC20Holders(rpcEndpoint, params.address, params.atBlock, options),
  },
  ERC721HolderForToken: {
    label: 'Holder for ERC721',
    value: 'ERC721HolderForToken',
    fields: {
      tokenId: 'required',
      block: 'optional',
    },
    View: TokenHolderTable,
    getItems: (rpcEndpoint, params, options) => getHolderForERC721Token(rpcEndpoint, params.address, params.tokenId, params.atBlock),
  },
  ERC721Holders: {
    label: 'ERC721 Token Holders',
    value: 'ERC721Holders',
    fields: {
      block: 'optional',
    },
    View: TokenHolderTable,
    getItems: (rpcEndpoint, params, options) => getERC721Holders(rpcEndpoint, params.address, params.atBlock, options),
  },
  ERC721Tokens: {
    label: 'ERC721 Tokens',
    value: 'ERC721Tokens',
    fields: {
      block: 'optional',
    },
    View: TokenTable,
    getItems: (rpcEndpoint, params, options) => getERC721Tokens(rpcEndpoint, params.address, params.atBlock, options),
  },
  ERC721TokensForAccount: {
    label: 'ERC721 Tokens for Account',
    value: 'ERC721TokensForAccount',
    fields: {
      account: 'required',
      block: 'optional',
    },
    View: TokenTable,
    getItems: (rpcEndpoint, params, options) => getERC721TokensForAccount(rpcEndpoint, params.address, params.account, params.atBlock, options),
  },
  Events: {
    label: 'Contract Events',
    value: 'Events',
    fields: {},
    View: EventTable,
    getItems: (rpcEndpoint, params, options) => getEvents(rpcEndpoint, params.address, options),
  },
  InternalToTxs: {
    label: 'Internal Transactions to Contract',
    value: 'InternalToTxs',
    fields: {},
    View: TransactionTable,
    getItems: (rpcEndpoint, params, options) => getInternalToTxs(rpcEndpoint, params.address, options),
  },
  ToTxs: {
    label: 'Transactions To Contract',
    value: 'ToTxs',
    fields: {},
    View: TransactionTable,
    getItems: (rpcEndpoint, params, options) => getToTxs(rpcEndpoint, params.address, options),
  },
}

export function getReportsForTemplate (templateName) {
  const commonReports = [Reports.ToTxs, Reports.InternalToTxs, Reports.Events]
  switch (templateName) {
    case 'ERC20':
      return [Reports.ERC20TokenHolders, Reports.ERC20TokenBalance, ...commonReports]
    case 'ERC721':
      return [Reports.ERC721Holders, Reports.ERC721Tokens, Reports.ERC721TokensForAccount, Reports.ERC721HolderForToken, ...commonReports]
    default:
      return [...commonReports, Reports.GenerateReport]
  }
}

export function getDefaultReportForTemplate (templateName) {
  return getReportsForTemplate(templateName)[0]
}

export default Reports
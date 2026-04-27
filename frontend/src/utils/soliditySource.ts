/**
 * 将 Etherscan Standard JSON Input（或已双重 JSON 字符串化的内容）展开为可读的 Solidity 多文件文本。
 * 用于漏洞/题目合约源码展示与下载；与后端 normalizeEtherscanSourceCode 行为对齐。
 */
export function formatSolidityOrJSONSource(raw: string): string {
  const trimmed = raw.trim()
  if (!trimmed) {
    return raw
  }
  let jsonStr = trimmed
  if (jsonStr.startsWith('{{') && jsonStr.endsWith('}}')) {
    jsonStr = jsonStr.slice(1, -1).trim()
  }
  try {
    const parsed = JSON.parse(jsonStr) as { sources?: Record<string, { content?: string }> }
    const sources = parsed?.sources
    if (!sources || typeof sources !== 'object') {
      return raw
    }
    const paths = Object.keys(sources).sort()
    if (paths.length === 0) {
      return raw
    }
    const parts: string[] = []
    for (const path of paths) {
      const content = sources[path]?.content ?? ''
      parts.push(`// ===== File: ${path} =====\n${content}${content.endsWith('\n') ? '' : '\n'}`)
    }
    return parts.join('\n')
  } catch {
    return raw
  }
}

export function isMultiFileSolidityBundle(formatted: string): boolean {
  return formatted.includes('// ===== File:')
}

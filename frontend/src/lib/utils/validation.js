// フォーム検証ユーティリティ

/**
 * 入力値のサニタイゼーション
 */
export function sanitizeInput(input) {
  if (typeof input !== 'string') {
    return input;
  }

  return input
    .trim()
    // HTMLタグを除去
    .replace(/<[^>]*>/g, '')
    // スクリプトタグを除去
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')
    // JavaScriptイベントハンドラーを除去
    .replace(/on\w+\s*=\s*["'][^"']*["']/gi, '')
    // javascript:プロトコルを除去
    .replace(/javascript:/gi, '')
    // 危険な文字をエスケープ
    .replace(/[<>&"']/g, (match) => {
      const escapeMap = {
        '<': '&lt;',
        '>': '&gt;',
        '&': '&amp;',
        '"': '&quot;',
        "'": '&#x27;'
      };
      return escapeMap[match];
    });
}

/**
 * HTMLエンティティのエスケープ
 */
export function escapeHtml(text) {
  if (typeof text !== 'string') {
    return text;
  }

  const escapeMap = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#x27;',
    '/': '&#x2F;'
  };

  return text.replace(/[&<>"'\/]/g, (match) => escapeMap[match]);
}

/**
 * SQLインジェクション対策のための文字列検証
 */
export function validateSqlSafety(input) {
  if (typeof input !== 'string') {
    return { isValid: true, error: null };
  }

  // 危険なSQLキーワードのパターン
  const dangerousPatterns = [
    /(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|EXECUTE|UNION|SCRIPT)\b)/gi,
    /(--|\/\*|\*\/|;|'|"|`)/g,
    /(\bOR\b|\bAND\b).*(\b=\b|\b<\b|\b>\b)/gi
  ];

  for (const pattern of dangerousPatterns) {
    if (pattern.test(input)) {
      return { 
        isValid: false, 
        error: '不正な文字が含まれています' 
      };
    }
  }

  return { isValid: true, error: null };
}

/**
 * XSS攻撃対策のための文字列検証
 */
export function validateXssSafety(input) {
  if (typeof input !== 'string') {
    return { isValid: true, error: null };
  }

  // 危険なXSSパターン
  const xssPatterns = [
    /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
    /<iframe\b[^<]*(?:(?!<\/iframe>)<[^<]*)*<\/iframe>/gi,
    /<object\b[^<]*(?:(?!<\/object>)<[^<]*)*<\/object>/gi,
    /<embed\b[^<]*(?:(?!<\/embed>)<[^<]*)*<\/embed>/gi,
    /on\w+\s*=\s*["'][^"']*["']/gi,
    /javascript:/gi,
    /vbscript:/gi,
    /data:text\/html/gi
  ];

  for (const pattern of xssPatterns) {
    if (pattern.test(input)) {
      return { 
        isValid: false, 
        error: '不正なスクリプトが含まれています' 
      };
    }
  }

  return { isValid: true, error: null };
}

/**
 * 文字列の長さ検証
 */
export function validateLength(value, min = 0, max = Infinity, fieldName = 'フィールド') {
  if (typeof value !== 'string') {
    return { isValid: false, error: `${fieldName}は文字列である必要があります` };
  }

  const length = value.trim().length;

  if (length < min) {
    return { 
      isValid: false, 
      error: `${fieldName}は${min}文字以上である必要があります` 
    };
  }

  if (length > max) {
    return { 
      isValid: false, 
      error: `${fieldName}は${max}文字以下である必要があります` 
    };
  }

  return { isValid: true, error: null };
}

/**
 * 数値の範囲検証
 */
export function validateRange(value, min = -Infinity, max = Infinity, fieldName = '値') {
  const numValue = Number(value);

  if (isNaN(numValue)) {
    return { 
      isValid: false, 
      error: `${fieldName}は数値である必要があります` 
    };
  }

  if (numValue < min) {
    return { 
      isValid: false, 
      error: `${fieldName}は${min}以上である必要があります` 
    };
  }

  if (numValue > max) {
    return { 
      isValid: false, 
      error: `${fieldName}は${max}以下である必要があります` 
    };
  }

  return { isValid: true, error: null };
}

/**
 * パターンマッチング検証
 */
export function validatePattern(value, pattern, errorMessage) {
  if (typeof value !== 'string') {
    return { isValid: false, error: '文字列である必要があります' };
  }

  if (!pattern.test(value)) {
    return { 
      isValid: false, 
      error: errorMessage || 'フォーマットが正しくありません' 
    };
  }

  return { isValid: true, error: null };
}

// ログイン情報の検証（強化版）
export function validateLoginCredentials(username, password) {
  const errors = {};

  // ユーザー名の検証
  if (!username || username.trim().length === 0) {
    errors.username = 'ユーザー名は必須です';
  } else {
    // セキュリティ検証（サニタイゼーション前の原文で）
    const sqlValidation = validateSqlSafety(username);
    if (!sqlValidation.isValid) {
      errors.username = sqlValidation.error;
    } else {
      const xssValidation = validateXssSafety(username);
      if (!xssValidation.isValid) {
        errors.username = xssValidation.error;
      } else {
        // サニタイゼーション
        const sanitizedUsername = sanitizeInput(username);
        
        // 長さ検証
        const lengthValidation = validateLength(sanitizedUsername, 1, 50, 'ユーザー名');
        if (!lengthValidation.isValid) {
          errors.username = lengthValidation.error;
        }
      }
    }
  }

  // パスワードの検証
  if (!password || password.length === 0) {
    errors.password = 'パスワードは必須です';
  } else {
    // 長さ検証
    const lengthValidation = validateLength(password, 1, 100, 'パスワード');
    if (!lengthValidation.isValid) {
      errors.password = lengthValidation.error;
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
    sanitizedData: {
      username: username ? sanitizeInput(username) : '',
      password: password || ''
    }
  };
}

// スコア検証（強化版）
export function validateScore(score, fieldName = 'スコア') {
  if (score === null || score === undefined || score === '') {
    return { isValid: false, error: `${fieldName}は必須です` };
  }

  // 文字列の場合はセキュリティ検証とサニタイゼーション
  let sanitizedScore = score;
  if (typeof score === 'string') {
    // セキュリティ検証（サニタイゼーション前の原文で）
    const sqlValidation = validateSqlSafety(score);
    if (!sqlValidation.isValid) {
      return { isValid: false, error: sqlValidation.error };
    }

    const xssValidation = validateXssSafety(score);
    if (!xssValidation.isValid) {
      return { isValid: false, error: xssValidation.error };
    }
    
    // サニタイゼーション
    sanitizedScore = sanitizeInput(score);
  }

  // 数値範囲検証
  const rangeValidation = validateRange(sanitizedScore, 0, 999, fieldName);
  if (!rangeValidation.isValid) {
    return rangeValidation;
  }

  return { 
    isValid: true, 
    error: null,
    sanitizedValue: Number(sanitizedScore)
  };
}

// 試合結果検証（強化版）
export function validateMatchResult(score1, score2, team1Name = '', team2Name = '') {
  const errors = {};
  const sanitizedData = {};

  // スコア1の検証
  const score1Validation = validateScore(score1, 'チーム1のスコア');
  if (!score1Validation.isValid) {
    errors.score1 = score1Validation.error;
  } else {
    sanitizedData.score1 = score1Validation.sanitizedValue;
  }

  // スコア2の検証
  const score2Validation = validateScore(score2, 'チーム2のスコア');
  if (!score2Validation.isValid) {
    errors.score2 = score2Validation.error;
  } else {
    sanitizedData.score2 = score2Validation.sanitizedValue;
  }

  // チーム名の検証（オプション）
  if (team1Name) {
    const team1Validation = validateTeamName(team1Name);
    if (!team1Validation.isValid) {
      errors.team1 = team1Validation.error;
    } else {
      sanitizedData.team1 = team1Validation.sanitizedValue;
    }
  }

  if (team2Name) {
    const team2Validation = validateTeamName(team2Name);
    if (!team2Validation.isValid) {
      errors.team2 = team2Validation.error;
    } else {
      sanitizedData.team2 = team2Validation.sanitizedValue;
    }
  }

  // 論理的検証
  if (sanitizedData.score1 !== undefined && sanitizedData.score2 !== undefined) {
    if (sanitizedData.score1 === sanitizedData.score2) {
      errors.general = '引き分けは許可されていません';
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
    sanitizedData
  };
}

// チーム名検証
export function validateTeamName(teamName) {
  if (!teamName || teamName.trim().length === 0) {
    return { isValid: false, error: 'チーム名は必須です' };
  }

  // セキュリティ検証（サニタイゼーション前の原文で）
  const sqlValidation = validateSqlSafety(teamName);
  if (!sqlValidation.isValid) {
    return sqlValidation;
  }

  const xssValidation = validateXssSafety(teamName);
  if (!xssValidation.isValid) {
    return xssValidation;
  }

  // サニタイゼーション
  const sanitizedName = sanitizeInput(teamName);

  // 長さ検証
  const lengthValidation = validateLength(sanitizedName, 1, 50, 'チーム名');
  if (!lengthValidation.isValid) {
    return lengthValidation;
  }

  return { 
    isValid: true, 
    error: null,
    sanitizedValue: sanitizedName
  };
}

// 汎用的な必須フィールド検証（強化版）
export function validateRequired(value, fieldName) {
  if (!value || (typeof value === 'string' && value.trim().length === 0)) {
    return { isValid: false, error: `${fieldName}は必須です` };
  }

  // 文字列の場合はセキュリティ検証も実行
  if (typeof value === 'string') {
    const sanitizedValue = sanitizeInput(value);
    
    const sqlValidation = validateSqlSafety(sanitizedValue);
    if (!sqlValidation.isValid) {
      return sqlValidation;
    }

    const xssValidation = validateXssSafety(sanitizedValue);
    if (!xssValidation.isValid) {
      return xssValidation;
    }

    return { 
      isValid: true, 
      error: null,
      sanitizedValue
    };
  }

  return { isValid: true, error: null };
}

/**
 * フォーム全体の検証
 */
export function validateForm(formData, validationRules) {
  const errors = {};
  const sanitizedData = {};

  for (const [fieldName, rules] of Object.entries(validationRules)) {
    const value = formData[fieldName];
    let fieldErrors = [];

    for (const rule of rules) {
      const result = rule.validator(value, rule.params);
      if (!result.isValid) {
        fieldErrors.push(result.error);
      } else if (result.sanitizedValue !== undefined) {
        sanitizedData[fieldName] = result.sanitizedValue;
      }
    }

    if (fieldErrors.length > 0) {
      errors[fieldName] = fieldErrors[0]; // 最初のエラーのみ表示
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
    sanitizedData
  };
}

/**
 * リアルタイム検証用のデバウンス関数
 */
export function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

/**
 * CSRFトークン検証（フロントエンド側）
 */
export function validateCsrfToken(token) {
  if (!token || typeof token !== 'string') {
    return { isValid: false, error: 'CSRFトークンが無効です' };
  }

  // トークンの形式検証（例：32文字の英数字）
  const tokenPattern = /^[a-zA-Z0-9]{32,}$/;
  if (!tokenPattern.test(token)) {
    return { isValid: false, error: 'CSRFトークンの形式が正しくありません' };
  }

  return { isValid: true, error: null };
}

// フォーム検証ユーティリティ

// ログイン情報の検証
export function validateLoginCredentials(username, password) {
  const errors = {};

  if (!username || username.trim().length === 0) {
    errors.username = 'ユーザー名は必須です';
  }

  if (!password || password.length === 0) {
    errors.password = 'パスワードは必須です';
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors
  };
}

// スコア検証
export function validateScore(score) {
  if (score === null || score === undefined || score === '') {
    return { isValid: false, error: 'スコアは必須です' };
  }

  const numScore = Number(score);
  if (isNaN(numScore) || numScore < 0) {
    return { isValid: false, error: 'スコアは0以上の数値である必要があります' };
  }

  return { isValid: true, error: null };
}

// 試合結果検証
export function validateMatchResult(score1, score2) {
  const score1Validation = validateScore(score1);
  const score2Validation = validateScore(score2);

  const errors = {};

  if (!score1Validation.isValid) {
    errors.score1 = score1Validation.error;
  }

  if (!score2Validation.isValid) {
    errors.score2 = score2Validation.error;
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors
  };
}

// 汎用的な必須フィールド検証
export function validateRequired(value, fieldName) {
  if (!value || (typeof value === 'string' && value.trim().length === 0)) {
    return { isValid: false, error: `${fieldName}は必須です` };
  }

  return { isValid: true, error: null };
}

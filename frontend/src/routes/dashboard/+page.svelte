<script>
	import { goto } from '$app/navigation';
	import { tick } from 'svelte';
	import { createBracket } from 'bracketry';
	import RelayTable from '../../lib/RelayTable.svelte';
	import AttendanceInput from '../../lib/AttendanceInput.svelte';
	import { dndzone } from 'svelte-dnd-action';

	// --- State Variables ---

	// General
	let userRole = '';
	let assignedSport = '';
	let activeTab = 'tournament'; // 'tournament' | 'input' | 'scores'
	let isLoading = false;
	let classNameMap = {};

	// 学年名のマッピング
	const gradeNames = {
		1: '1年生',
		2: '2年生', 
		3: '3年生',
		4: '4年生',
		5: '5年生',
		6: '専・教'
	};

	// Tournament
	let allTournaments = [];
	let selectedTournament = null;
	let bracketContainer;
	let selectedSport = '';
	let tableTennisWeather = 'sunny';
	let isRainyChecked = false;
	$: isRainyChecked = tableTennisWeather === 'rainy';

	// Match Input
	let matchesBySport = {};
	let matchesLoading = false;
	let showConfirmModal = false;
	let editingMatch = null;
	let editingSport = '';
	let showResetConfirmModal = false;
	let resettingMatch = null;
	let activeInputTab = 'tournament'; // 'tournament' | 'relay'
	let showWinnerModal = false;
	let selectedWinnerId = null;

	// Score
	let scores = [];
	let scoresLoading = false;
	let showTotalScores = true; // Default value
	let showQuestionnaireButton = false;

	// Attendance
	let attendanceScores = [];
	let attendanceLoading = false;
	let attendanceUpdateStatus = '';

	// Relay
	let relayActive = false; // For tournament tab's relay view
	let relayType = 'A';
	let relayResults = [];
	let relayLoading = false;
	let relayError = '';
	let relayUpdateStatus = ''; // 'success' | 'error' | ''
	let showRelayResetConfirmModal = false;
	let resettingRelayType = '';


	// --- Lifecycle & Initialisation ---

	import { onMount } from 'svelte';
	onMount(async () => {
		await fetchUser();
		await fetchVisibility();
		await fetchWeather();
		await fetchTournament('volleyball');
		await fetchClassNameMap();
		if (userRole === 'superroot') {
			await fetchAttendanceScores();
		}
	});

	async function fetchUser() {
		try {
			const res = await fetch('/api/auth/me', {
				credentials: 'include'
			});
			if (res.ok) {
				const user = await res.json();
				userRole = user.role || '';
				assignedSport = user.assigned_sport || '';
			} else {
				goto('/login');
			}
		} catch (e) {
			console.error('Failed to fetch user', e);
			goto('/login');
		}
	}

	async function logout() {
		try {
			await fetch('/api/auth/logout', {
				method: 'POST',
				credentials: 'include'
			});
		} catch (e) {
			console.error('Logout failed', e);
		}
		goto('/login');
	}


	// --- Data Fetching ---

	async function fetchVisibility() {
		try {
			const res = await fetch('/api/settings/visibility', {
				credentials: 'include'
			});
			if (res.ok) {
				const data = await res.json();
				showTotalScores = data.showTotalScores;
				showQuestionnaireButton = data.showQuestionnaireButton;
			}
		} catch (e) {
			console.error('Failed to fetch visibility settings', e);
		}
	}

	async function updateVisibility(newValue) {
		try {
			await fetch('/api/settings/visibility', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ showTotalScores: newValue })
			});
		} catch (e) {
			console.error('Failed to update visibility settings', e);
			// Optionally, revert the UI change
			showTotalScores = !newValue;
		}
	}

	async function updateQuestionnaireButtonVisibility(newValue) {
		try {
			await fetch('/api/settings/visibility', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ showQuestionnaireButton: newValue })
			});
		} catch (e) {
			console.error('Failed to update questionnaire button visibility settings', e);
			showQuestionnaireButton = !newValue;
		}
	}

	async function fetchWeather() {
		try {
			const res = await fetch('/api/settings/weather', {
				credentials: 'include'
			});
			if (res.ok) {
				const data = await res.json();
				tableTennisWeather = data.tableTennisWeather;
			}
		} catch (e) {
			console.error('Failed to fetch weather settings', e);
		}
	}

	async function updateWeather(newValue) {
		try {
			await fetch('/api/settings/weather', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ tableTennisWeather: newValue })
			});
		} catch (e) {
			console.error('Failed to update weather settings', e);
		}
	}

	async function fetchTournament(sport) {
		selectedSport = sport;
		isLoading = true;
		allTournaments = [];
		selectedTournament = null;
		try {
			let url = `/api/tournaments/${sport}`;
			const res = await fetch(url, {
				credentials: 'include'
			});
			if (res.ok) {
				const data = await res.json();
				if (data && data.length > 0) {
					allTournaments = data;
					if (sport === 'table_tennis') {
						const filtered = getFilteredTableTennisTournaments();
						selectTournament(filtered[0]);
					} else {
						selectTournament(allTournaments[0]);
					}
				} else {
					allTournaments = [];
				}
			} else {
				alert('トーナメント情報の取得に失敗しました。');
			}
		} catch (error) {
			console.error('Fetch error:', error);
			alert('サーバーとの通信に失敗しました。');
		} finally {
			isLoading = false;
		}
	}

	async function fetchMatches(sport) {
		matchesLoading = true;
		try {
			let url = `/api/matches/${sport}`;
			if (sport === 'table_tennis') {
				url += `?weather=${tableTennisWeather}`;
			}
			const res = await fetch(url, {
				credentials: 'include'
			});
			if (res.ok) {
				matchesBySport[sport] = await res.json();
			} else {
				matchesBySport[sport] = [];
			}
		} catch (e) {
			matchesBySport[sport] = [];
		} finally {
			matchesLoading = false;
		}
	}

	async function fetchScores() {
		scoresLoading = true;
		const classOrder = ['1-1', '1-2', '1-3', 'IS2', 'IT2', 'IE2', 'IS3', 'IT3', 'IE3', 'IS4', 'IT4', 'IE4', 'IS5', 'IT5', 'IE5', '専・教'];
		const questionnaireScores = {
            '1-1': 5, '1-2': 5, '1-3': 5,
            'IS2': 4, 'IT2': 8, 'IE2': 2,
            'IS3': 5, 'IT3': 5, 'IE3': 5,
            'IS4': 9, 'IT4': 2, 'IE4': 4,
            'IS5': 2, 'IT5': 1, 'IE5': 4
        };
		try {
			const res = await fetch('/api/score', {
				credentials: 'include'
			});
			if (res.ok) {
				let fetchedScores = await res.json();
				fetchedScores = fetchedScores.map(s => {
                    const questionnaire_score = questionnaireScores[s.class_name] || 0;
                    return {
                        ...s,
                        questionnaire_score: questionnaire_score,
                        total_including_init: s.total_including_init + questionnaire_score
                    };
                });
				fetchedScores.sort((a, b) => {
                    const indexA = classOrder.indexOf(a.class_name);
                    const indexB = classOrder.indexOf(b.class_name);
                    if (indexA === -1) return 1;
                    if (indexB === -1) return -1;
                    return indexA - indexB;
                });
				scores = fetchedScores;
			}
		} catch (e) {
			scores = [];
		} finally {
			scoresLoading = false;
		}
	}

	async function fetchRelayResults(type) {
		relayLoading = true;
		relayError = '';
		relayType = type;
		try {
			const res = await fetch(`/api/relay?block=${type}`, {
				credentials: 'include'
			});
			if (res.ok) {
				let data = await res.json();
				if (data && data.rankings && Object.keys(data.rankings).length > 0) {
					const formattedData = Object.entries(data.rankings)
						.sort(([rankA], [rankB]) => parseInt(rankA) - parseInt(rankB))
						.map(([rank, grade]) => ({
							id: `relay-${type}-${rank}-${grade}-${Date.now()}`,
							rank: parseInt(rank),
							grade: parseInt(grade),
							relay_type: type
						}));
					relayResults = formattedData;
				} else {
					if (userRole === 'superroot' || (userRole === 'admin' && assignedSport === 'relay')) {
						relayResults = [1, 2, 3, 4, 5, 6].map(rank => ({
							id: `relay-${type}-${rank}-${Date.now()}`,
							rank: rank,
							grade: rank,
							relay_type: type
						}));
					} else {
						relayResults = [1, 2, 3, 4, 5, 6].map(rank => ({
							id: `relay-${type}-${rank}-${Date.now()}`,
							rank: rank,
							grade: null,
							relay_type: type
						}));
					}
				}
			} else {
				relayError = 'リレー結果の取得に失敗しました';
			}
		} catch (e) {
			console.error("Fetch error:", e);
			relayError = 'サーバー通信エラー、またはデータの形式が不正です。';
		} finally {
			relayLoading = false;
		}
	}

	async function fetchClassNameMap() {
		try {
			const res = await fetch('/api/score', { credentials: 'include' });
			if (res.ok) {
				const scoresData = await res.json();
				classNameMap = scoresData.reduce((acc, s) => {
					acc[s.class_id] = s.class_name;
					return acc;
				}, {});
			}
		} catch {}
	}


	// --- UI Logic & Event Handlers ---

	function handleInputTabClick() {
		activeTab = 'input';
		if (userRole === 'admin' && assignedSport === 'relay') {
			activeInputTab = 'relay';
			fetchRelayResults(relayType);
		} else {
			activeInputTab = 'tournament';
			const sports = userRole === 'superroot' ? ['volleyball', 'table_tennis', 'soccer'] : (assignedSport ? [assignedSport] : []);
			sports.forEach(fetchMatches);
		}
	}


	function getFilteredTableTennisMatches() {
		const allTableTennisMatches = matchesBySport['table_tennis'] ?? [];
		if (tableTennisWeather === 'sunny') {
			return allTableTennisMatches.filter(m => m.tournament_name === '卓球（晴天時）');
		}
		return allTableTennisMatches.filter(m => m.tournament_name.includes('雨天時'));
	}

	function getFilteredTableTennisTournaments() {
		if (!allTournaments) return [];
		return tableTennisWeather === 'sunny'
			? allTournaments.filter(t => t.name === '卓球（晴天時）')
			: allTournaments.filter(t => t.name.includes('雨天時'));
	}

	function selectTournament(tournamentData) {
		selectedTournament = tournamentData;
		tick().then(drawBracket);
	}

	async function drawBracket() {
		if (!selectedTournament || !bracketContainer) return;
		bracketContainer.innerHTML = '';
		await tick();
		createBracket(selectedTournament, bracketContainer);
	}

	function handleSubmit(match, sport) {
		editingMatch = match;
		editingSport = sport;
		const score1 = Number(match.team1_score);
		const score2 = Number(match.team2_score);

		if (score1 === score2) {
			selectedWinnerId = match.team1_id; // Default selection
			showWinnerModal = true;
		} else {
			showConfirmModal = true;
		}
	}

	function confirmWinnerSelection() {
		if (editingMatch) {
			editingMatch.winner_team_id = parseInt(selectedWinnerId, 10);
		}
		showWinnerModal = false;
		showConfirmModal = true; // Chain to the next modal
	}


	// --- API Updates ---

	async function updateMatchScore(match, sport) {
		const body = {
			team1_score: Number(match.team1_score),
			team2_score: Number(match.team2_score),
		};

		if (Number(match.team1_score) === Number(match.team2_score) && match.winner_team_id) {
			body.winner_team_id = match.winner_team_id;
		}

		try {
			const res = await fetch(`/api/matches/${match.id}`, {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify(body)
			});
			match._updateStatus = res.ok ? 'success' : 'error';
		} catch (e) {
			match._updateStatus = 'error';
		}
		setTimeout(() => { match._updateStatus = ''; }, 2000);
		await fetchMatches(sport);
		if (sport === 'table_tennis') {
			matchesBySport['table_tennis'] = [...(matchesBySport['table_tennis'] ?? [])];
		}
	}

	function confirmUpdateMatchScore() {
		if (editingMatch && editingSport) {
			updateMatchScore(editingMatch, editingSport);
			location.reload();
		}
		showConfirmModal = false;
		editingMatch = null;
		editingSport = '';
	}

	function confirmResetMatch() {
		if (resettingMatch && editingSport) {
			resetMatch(resettingMatch, editingSport);
			location.reload();
		}
		showResetConfirmModal = false;
		resettingMatch = null;
		editingSport = '';
	}

	async function resetMatch(match, sport) {
		try {
			const res = await fetch(`/api/matches/${match.id}/reset`, {
				method: 'POST',
				credentials: 'include'
			});
			match._updateStatus = res.ok ? 'success' : 'error';
		} catch (e) {
			match._updateStatus = 'error';
		}
		setTimeout(() => { match._updateStatus = ''; }, 2000);
		await fetchMatches(sport);
		if (sport === 'table_tennis') {
			matchesBySport['table_tennis'] = [...(matchesBySport['table_tennis'] ?? [])];
		}
	}

	function handleDndConsider(event) {
		if (event.detail.items && Array.isArray(event.detail.items)) {
			relayResults = event.detail.items.map((item, index) => ({
				...item,
				rank: index + 1
			}));
		}
	}

	function handleDndFinalize(event) {
		if (event.detail.items && Array.isArray(event.detail.items)) {
			relayResults = event.detail.items.map((item, index) => ({
				...item,
				rank: index + 1
			}));
		}
	}

	async function updateRelayRanks() {
		relayUpdateStatus = 'loading';

		if (relayResults.length !== 6) {
			alert('リレーのクラス数が6ではありません。');
			relayUpdateStatus = '';
			return;
		}

		const rankings = {};
		relayResults.forEach((result, index) => {
			rankings[index + 1] = result.grade;
		});

		const body = { rankings };

		try {
			const res = await fetch(`/api/relay?block=${relayType}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify(body)
			});
			if (res.ok) {
				relayUpdateStatus = 'success';
				await fetchRelayResults(relayType);
			} else {
				const errorData = await res.json();
				alert(`更新に失敗しました: ${errorData.error}`);
				relayUpdateStatus = 'error';
			}
		} catch (e) {
			console.error('Relay update error:', e);
			alert('サーバーとの通信に失敗しました。');
			relayUpdateStatus = 'error';
		}
		setTimeout(() => { relayUpdateStatus = ''; }, 3000);
	}

	function openRelayResetModal(block) {
		resettingRelayType = block;
		showRelayResetConfirmModal = true;
	}

	async function handleRelayReset() {
		if (!resettingRelayType) return;

		relayUpdateStatus = 'loading';
		try {
			const res = await fetch('/api/relay/reset', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ block: resettingRelayType })
			});

			if (res.ok) {
				relayUpdateStatus = 'success';
				await fetchRelayResults(resettingRelayType);
			} else {
				const errorData = await res.json();
				alert(`リセットに失敗しました: ${errorData.error}`);
				relayUpdateStatus = 'error';
			}
		} catch (e) {
			console.error('Relay reset error:', e);
			alert('サーバーとの通信に失敗しました。');
			relayUpdateStatus = 'error';
		}

		showRelayResetConfirmModal = false;
		resettingRelayType = '';
		setTimeout(() => { relayUpdateStatus = ''; }, 3000);
	}

	// 出席点関連の関数
	async function fetchAttendanceScores() {
		attendanceLoading = true;
		try {
			const res = await fetch('/api/attendance', {
				credentials: 'include'
			});
			if (res.ok) {
				attendanceScores = await res.json();
			} else {
				console.error('Failed to fetch attendance scores');
			}
		} catch (e) {
			console.error('Error fetching attendance scores:', e);
		} finally {
			attendanceLoading = false;
		}
	}

	function handleAttendanceScoreChange(event) {
		const { classId, score } = event.detail;
		attendanceScores = attendanceScores.map(item => 
			item.class_id === classId ? { ...item, score } : item
		);
	}

	async function handleAttendanceBatchUpdate() {
		attendanceUpdateStatus = 'loading';
		
		const scores = attendanceScores.map(item => ({
			class_id: item.class_id,
			score: item.score
		}));

		try {
			const res = await fetch('/api/attendance/batch', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify({ scores })
			});

			if (res.ok) {
				attendanceUpdateStatus = 'success';
				await fetchScores();
			} else {
				attendanceUpdateStatus = 'error';
				const errorData = await res.json();
				alert(`更新に失敗しました: ${errorData.error}`);
			}
		} catch (e) {
			console.error('Error updating attendance scores:', e);
			attendanceUpdateStatus = 'error';
			alert('サーバーとの通信に失敗しました。');
		}

		setTimeout(() => { attendanceUpdateStatus = ''; }, 3000);
	}

	const scoreCategories = [
        { name: '春スポ体合計点', key: 'init_score' },
        { name: '春スポアンケート点', key: 'questionnaire_score' },
        { name: '出席点', key: 'attendance_score' },
        { name: 'バレーボール1勝点', key: 'volleyball1_score' },
        { name: 'バレーボール2勝点', key: 'volleyball2_score' },
        { name: 'バレーボール3勝点', key: 'volleyball3_score' },
        { name: 'バレーボール優勝点', key: 'volleyball_championship_score' },
        { name: '卓球1勝点', key: 'table_tennis1_score' },
        { name: '卓球2勝点', key: 'table_tennis2_score' },
        { name: '卓球3勝点', key: 'table_tennis3_score' },
        { name: '卓球優勝点', key: 'table_tennis_championship_score' },
        { name: '卓球雨天ボーナス', key: 'table_tennis_rainy_bonus_score' },
        { name: 'サッカー1勝点', key: 'soccer1_score' },
        { name: 'サッカー2勝点', key: 'soccer2_score' },
        { name: 'サッカー3勝点', key: 'soccer3_score' },
        { name: 'サッカー優勝点', key: 'soccer_championship_score' },
        { name: 'リレーAブロック得点', key: 'relay_A_score' },
        { name: 'リレーBブロック得点', key: 'relay_B_score' },
        { name: 'リレーボーナス得点', key: 'relay_bonus_score' },
        { name: '合計(春スポ体合計点除く)', key: 'total_excluding_init' },
        { name: '今回の順位', key: 'current_rank_excluding_init' },
        { name: '合計(春スポ体合計点含む)', key: 'total_including_init' },
        { name: '総合順位', key: 'current_rank' },
    ];

	function calculateRanks(scoresData) {
    	if (!scoresData || scoresData.length === 0) {
        	return [];
    	}
    	const sortedByTotal = [...scoresData].sort((a, b) => b.total_including_init - a.total_including_init);
    	const rankMapTotal = new Map();
    	let rankTotal = 1;
    	for (let i = 0; i < sortedByTotal.length; i++) {
        	if (i > 0 && sortedByTotal[i].total_including_init < sortedByTotal[i - 1].total_including_init) {
            	rankTotal = i + 1;
        	}
        	rankMapTotal.set(sortedByTotal[i].class_name, rankTotal);
    	}

    	const sortedByEvent = [...scoresData].sort((a, b) => b.total_excluding_init - a.total_excluding_init);
    	const rankMapEvent = new Map();
    	let rankEvent = 1;
    	for (let i = 0; i < sortedByEvent.length; i++) {
        	if (i > 0 && sortedByEvent[i].total_excluding_init < sortedByEvent[i - 1].total_excluding_init) {
            	rankEvent = i + 1;
        	}
        	rankMapEvent.set(sortedByEvent[i].class_name, rankEvent);
    	}

    	return scoresData.map(s => ({
        	...s,
        	current_rank: rankMapTotal.get(s.class_name),
			current_rank_excluding_init: rankMapEvent.get(s.class_name),
    	}));
	}

	$: scoresWithRanks = scores.length > 0 ? calculateRanks(scores) : [];

</script>

<svelte:head>
	<title>行事委員会結果速報アプリ</title>
</svelte:head>

<div class="dashboard-container">
	<header class="dashboard-header">
		<h1 class="dashboard-title">ダッシュボード</h1>
		<nav class="dashboard-tabs">
			<button class:active={activeTab === 'tournament'} on:click={() => { activeTab = 'tournament'; fetchTournament('volleyball'); }} class="dashboard-tab-btn">競技トーナメント</button>
			<button class:active={activeTab === 'scores'} on:click={() => { activeTab = 'scores'; fetchScores(); }} class="dashboard-tab-btn">現在の得点</button>
			{#if userRole === 'superroot' || userRole === 'admin' || userRole === 'admin_relay'}
				<button class:active={activeTab === 'input'} on:click={handleInputTabClick} class="dashboard-tab-btn">試合結果入力</button>
			{/if}
			{#if showQuestionnaireButton}
				<a href="https://forms.cloud.microsoft/Pages/ResponsePage.aspx?id=XYP-cpVeEkWK4KezivJfyAklshGgL0hOqHb1R7C4GBpUQVFBQUtPMDNMRzJOOEJQU0c3R0xDMlFXMi4u" target="_blank" rel="noopener noreferrer" class="dashboard-tab-btn">振り返りアンケート</a>
			{/if}
		</nav>
		<button class="logout-btn" on:click={logout}>ログアウト</button>
	</header>

	<main>
		{#if activeTab === 'tournament'}
			<div class="sports-buttons">
				<button on:click={() => { relayActive = false; fetchTournament('volleyball'); }} class:active={selectedSport === 'volleyball' && !relayActive} class="sports-btn">バレーボール</button>
				<button on:click={() => { relayActive = false; fetchTournament('table_tennis'); }} class:active={selectedSport === 'table_tennis' && !relayActive} class="sports-btn">卓球</button>
				<button on:click={() => { relayActive = false; fetchTournament('soccer'); }} class:active={selectedSport === 'soccer' && !relayActive} class="sports-btn">サッカー</button>
				<button on:click={() => { relayActive = true; fetchRelayResults('A'); }} class:active={relayActive} class="sports-btn">リレー</button>
				{#if (userRole === 'superroot' || (userRole === 'admin' && assignedSport === 'table_tennis')) && selectedSport === 'table_tennis' && !relayActive}
				<div class="weather-switcher">
					<span>卓球トーナメント天候:</span>
					<label class="switch">
						<input type="checkbox" checked={isRainyChecked} on:change={() => {
							tableTennisWeather = tableTennisWeather === 'sunny' ? 'rainy' : 'sunny';
							updateWeather(tableTennisWeather);
							fetchTournament('table_tennis');
						}}>
						<span class="slider"></span>
					</label>
					<span>{tableTennisWeather === 'sunny' ? '晴天' : '雨天'}</span>
				</div>
				{/if}
			</div>
			{#if relayActive}
				<div class="relay-area">
					<h2>学年対抗リレー結果</h2>
					<div class="relay-info">
						<p><strong>得点システム:</strong> 1位30点、2位25点、3位20点、4位15点、5位10点、6位5点</p>
						<p><strong>ボーナス得点:</strong> 両ブロック合計で1位30点、2位20点、3位10点の追加得点</p>
					</div>
					<div class="relay-type-selector">
						<button on:click={() => fetchRelayResults('A')} class:active={relayType === 'A'}>リレーAブロック</button>
						<button on:click={() => fetchRelayResults('B')} class:active={relayType === 'B'}>リレーBブロック</button>
					</div>
					{#if relayLoading}
						<p>読み込み中...</p>
					{:else if relayError}
						<p style="color:red">{relayError}</p>
					{:else}
						<RelayTable {relayResults} {relayType} {classNameMap} />
					{/if}
				</div>
			{:else}
				{#if isLoading}<p>読み込み中...</p>{/if}
				{#if !isLoading && selectedSport === 'table_tennis'}
					{#if getFilteredTableTennisTournaments().length > 1}
						<div class="tournament-selector">
							{#each getFilteredTableTennisTournaments() as t}
								<button on:click={() => selectTournament(t)} class:active={selectedTournament?.name === t.name}>{t.name}</button>
							{/each}
						</div>
					{/if}
				{:else if !isLoading && allTournaments.length > 1}
					<div class="tournament-selector">
						{#each allTournaments as t}
							<button on:click={() => selectTournament(t)} class:active={selectedTournament?.name === t.name}>{t.name}</button>
						{/each}
					</div>
				{/if}
				<div class="bracket-area">
					{#if selectedTournament}
						<h2>{selectedTournament.name} トーナメント表</h2>
						<div bind:this={bracketContainer} class="bracket-wrapper"></div>
					{/if}
				</div>
			{/if}

		{:else if activeTab === 'scores'}
    	<div class="scores-area">
        	<h2>現在の得点</h2>
        
        	{#if userRole === 'superroot'}
            	<div class="visibility-switcher">
                	<span>合計・順位の表示:</span>
                	<label class="switch">
                    	<input type="checkbox" bind:checked={showTotalScores} on:change={() => updateVisibility(showTotalScores)}>
                    	<span class="slider"></span>
                	</label>
                	<span>{showTotalScores ? '表示中' : '非表示'}</span>
            	</div>
				<div class="visibility-switcher">
					<span>振り返りアンケートボタンの表示:</span>
					<label class="switch">
						<input type="checkbox" bind:checked={showQuestionnaireButton} on:change={() => updateQuestionnaireButtonVisibility(showQuestionnaireButton)}>
						<span class="slider"></span>
					</label>
					<span>{showQuestionnaireButton ? '表示中' : '非表示'}</span>
				</div>
        	{/if}
        	{#if scoresLoading}
            	<p>読み込み中...</p>
        	{:else if scores.length > 0}
            	<div class="scores-container">
                	<div class="score-category-column">
                    	<div class="score-header">得点項目</div>
                    	{#each scoreCategories as category, i}
                        	{#if showTotalScores  || !['total_excluding_init', 'total_including_init', 'current_rank', 'current_rank_excluding_init'].includes(category.key)}
                            	<div class="score-cell" class:odd-row={i % 2 === 0}><b>{category.name}</b></div>
                        	{/if}
                    	{/each}
                	</div>
                	<div class="scores-data-wrapper">
                    	{#each scoresWithRanks as s}
                        	<div class="score-column">
                            	<div class="score-header">{s.class_name}</div>
                            	{#each scoreCategories as category, i}
                                	{#if showTotalScores  || !['total_excluding_init', 'total_including_init', 'current_rank', 'current_rank_excluding_init'].includes(category.key)}
                                    	<div class="score-cell" class:odd-row={i % 2 === 0} class:rank-cell={category.key.includes('_rank')}>
                                        	{#if category.key.includes('_rank')}
                                            	<span class="rank-badge rank-{s[category.key]}">{s[category.key]}位</span>
                                        	{:else}
                                            	{s[category.key]}
                                        	{/if}
                                    	</div>
                                	{/if}
                            	{/each}
                	        </div>
                    	{/each}
                	</div>
            	</div>
        	{:else}
            	<p>スコアデータがありません。</p>
        	{/if}
    	</div>

		{:else if activeTab === 'input'}
			<div class="match-input-area">
				<h2>試合結果入力</h2>
				
				<div class="input-tabs">
					{#if userRole === 'superroot' || userRole === 'admin'}
						<button class:active={activeInputTab === 'tournament'} on:click={() => activeInputTab = 'tournament'}>トーナメント</button>
					{/if}
					{#if userRole === 'superroot' || assignedSport === 'relay'}
						<button class:active={activeInputTab === 'relay'} on:click={() => { activeInputTab = 'relay'; fetchRelayResults(relayType); }}>リレー</button>
					{/if}
					{#if userRole === 'superroot'}
						<button class:active={activeInputTab === 'attendance'} on:click={() => activeInputTab = 'attendance'}>出席点</button>
					{/if}
				</div>

				{#if activeInputTab === 'tournament'}
					{#if userRole === 'admin'}<p>あなたの担当競技: <b>{assignedSport}</b></p>{/if}
					{#if userRole === 'superroot'}<p>全競技の試合結果を編集できます</p>{/if}
					{#if matchesLoading}<p>試合データ取得中...</p>{/if}

					{#each (userRole === 'superroot' ? ['volleyball', 'table_tennis', 'soccer'] : (assignedSport ? [assignedSport] : [])) as sport}
						<div class="tournament-edit-card">
							<h3>{sport} の試合一覧</h3>
							<div class="matches-list">
								{#if sport === 'table_tennis'}
									{#each getFilteredTableTennisMatches() as m (m.id)}
										<div class="match-entry">
											{#if m.status !== 'finished' || userRole === 'superroot'}
												<form class="match-edit-form" on:submit|preventDefault={() => handleSubmit(m, sport)}>
													<div class="match-info">
														<span>ID: {m.id}</span>
														<span>R: {m.round}</span>
														<span>{m.tournament_name}</span>
														<span>{m.team1_name || '-'} vs {m.team2_name || '-'}</span>
													</div>
													<div class="score-inputs">
														<span>{m.team1_name}</span>
														<input type="number" min="0" bind:value={m.team1_score} class="score-input" required/>
														<span> - </span>
														<input type="number" min="0" bind:value={m.team2_score} class="score-input" required />
														<span>{m.team2_name}</span>
													</div>
													<button type="submit" class="update-btn" disabled={!m.team1_name || !m.team2_name}>更新</button>
													{#if userRole === 'superroot'}
														<button type="button" class="reset-btn" on:click={() => { showResetConfirmModal = true; resettingMatch = m; editingSport = sport; }}>リセット</button>
													{/if}
													{#if m._updateStatus}
														<span class="update-status {m._updateStatus}">{m._updateStatus === 'success' ? '✓' : '✗'}</span>
													{/if}
												</form>
											{:else}
												<div class="match-info finished">
													<span>ID: {m.id}</span>
													<span>R: {m.round}</span>
													<span>{m.tournament_name}</span>
													<span>{m.team1_name} {m.team1_score} - {m.team2_score} {m.team2_name}</span>
													<span class="update-status success">試合終了</span>
													{#if userRole === 'superroot'}
														<button type="button" class="reset-btn" on:click={() => { showResetConfirmModal = true; resettingMatch = m; editingSport = sport; }}>リセット</button>
													{/if}
												</div>
											{/if}
										</div>
									{:else}
										<p>試合データがありません。</p>
									{/each}
								{:else}
									{#each matchesBySport[sport] as m (m.id)}
										<div class="match-entry">
											{#if m.status !== 'finished' || userRole === 'superroot'}
												<form class="match-edit-form" on:submit|preventDefault={() => handleSubmit(m, sport)}>
													<div class="match-info">
														<span>ID: {m.id}</span>
														<span>R: {m.round}</span>
														<span>{m.team1_name || '-'} vs {m.team2_name || '-'}</span>
													</div>
													<div class="score-inputs">
														<span>{m.team1_name}</span>
														<input type="number" min="0" bind:value={m.team1_score} class="score-input" required/>
														<span> - </span>
														<input type="number" min="0" bind:value={m.team2_score} class="score-input" required />
														<span>{m.team2_name}</span>
													</div>
													<button type="submit" class="update-btn" disabled={!m.team1_name || !m.team2_name}>更新</button>
													{#if userRole === 'superroot'}
														<button type="button" class="reset-btn" on:click={() => { showResetConfirmModal = true; resettingMatch = m; editingSport = sport; }}>リセット</button>
													{/if}
													{#if m._updateStatus}
														<span class="update-status {m._updateStatus}">{m._updateStatus === 'success' ? '✓' : '✗'}</span>
													{/if}
												</form>
											{:else}
												<div class="match-info finished">
													<span>ID: {m.id}</span>
													<span>R: {m.round}</span>
													<span>{m.tournament_name}</span>
													<span>{m.team1_name} {m.team1_score} - {m.team2_score} {m.team2_name}</span>
													<span class="update-status success">試合終了</span>
													{#if userRole === 'superroot'}
														<button type="button" class="reset-btn" on:click={() => { showResetConfirmModal = true; resettingMatch = m; editingSport = sport; }}>リセット</button>
													{/if}
												</div>
											{/if}
										</div>
									{:else}
										<p>試合データがありません。</p>
									{/each}
								{/if}
							</div>
						</div>
					{/each}

				{:else if activeInputTab === 'relay'}
					<div class="relay-input-card">
						<h3>リレー結果入力 (ドラッグ&ドロップで順位変更)</h3>
						<div class="relay-type-selector">
							<button on:click={() => fetchRelayResults('A')} class:active={relayType === 'A'}>リレーA</button>
							<button on:click={() => fetchRelayResults('B')} class:active={relayType === 'B'}>リレーB</button>
						</div>
						{#if relayLoading}
							<p>読み込み中...</p>
						{:else if relayError}
							<p style="color:red;">{relayError}</p>
						{:else if relayResults.length > 0}
							<section 
								class="dnd-list" 
								use:dndzone={{
									items: relayResults, 
									flipDurationMs: 300, 
									dropTargetStyle: {},
									dragDisabled: relayLoading || relayUpdateStatus === 'loading'
								}} 
								on:consider={handleDndConsider} 
								on:finalize={handleDndFinalize}
							>
								{#each relayResults as result, i (result.id)}
									<div class="dnd-item" data-id={result.id}>
										<span class="rank-badge">{result.rank || i + 1}位</span>
										<span class="class-name">{gradeNames[result.grade] || `学年${result.grade}`}</span>
									</div>
								{/each}
							</section>
							<div class="relay-actions">
								<button 
									on:click={updateRelayRanks} 
									class="update-btn" 
									disabled={relayUpdateStatus === 'loading'}
								>
									{relayUpdateStatus === 'loading' ? '更新中...' : 'リレー結果を更新'}
								</button>
								{#if userRole === 'superroot'}
									<button on:click={() => openRelayResetModal(relayType)} class="reset-btn">リセット</button>
								{/if}
							</div>
							{#if relayUpdateStatus === 'success'}
								<span class="update-status success">更新しました！</span>
							{:else if relayUpdateStatus === 'error'}
								<span class="update-status error">失敗しました。</span>
							{:else if relayUpdateStatus === 'loading'}
								<span class="update-status loading">処理中...</span>
							{/if}
						{:else}
							<p>リレーデータがありません。</p>
						{/if}
					</div>

				{:else if activeInputTab === 'attendance'}
					<AttendanceInput 
						{attendanceScores}
						loading={attendanceLoading}
						updateStatus={attendanceUpdateStatus}
						on:scoreChange={handleAttendanceScoreChange}
						on:batchUpdate={handleAttendanceBatchUpdate}
					/>
				{/if}
			</div>
		{/if}
	</main>

	{#if showConfirmModal}
    	<div class="modal-overlay" on:click={() => showConfirmModal = false}>
        	<div class="modal-content" on:click|stopPropagation>
            	<h3>試合結果を更新しますか？</h3>
				<p>ID: {editingMatch.id}, R: {editingMatch.round}</p>
				<p>{editingMatch.team1_name || '-'} vs {editingMatch.team2_name || '-'} </p>
				<p>スコア: {editingMatch.team1_score} - {editingMatch.team2_score}</p>
            	<p>本当にこの内容で更新してよろしいですか？</p>
            	<div class="modal-actions">
                	<button on:click={() => { showConfirmModal = false; editingMatch = null; editingSport = ''; }}>キャンセル</button>
					<button on:click={confirmUpdateMatchScore} class="ok-btn">OK</button>
            	</div>
        	</div>
    	</div>
	{/if}

	{#if showWinnerModal}
		<div class="modal-overlay" on:click={() => showWinnerModal = false}>
			<div class="modal-content" on:click|stopPropagation>
				<h3>勝者を選択してください</h3>
				<p>スコアが同点のため、勝者をマニュアルで選択する必要があります。</p>
				{#if editingMatch}
					<p><b>{editingMatch.team1_name || '-'}</b> vs <b>{editingMatch.team2_name || '-'}</b></p>
					<p>スコア: {editingMatch.team1_score} - {editingMatch.team2_score}</p>
					<div class="winner-select-container">
						<label for="winner-select">勝者:</label>
						<select id="winner-select" bind:value={selectedWinnerId}>
							<option value={editingMatch.team1_id}>{editingMatch.team1_name}</option>
							<option value={editingMatch.team2_id}>{editingMatch.team2_name}</option>
						</select>
					</div>
				{/if}
				<div class="modal-actions">
					<button on:click={() => { showWinnerModal = false; editingMatch = null; editingSport = ''; }}>キャンセル</button>
					<button on:click={confirmWinnerSelection} class="ok-btn">OK</button>
				</div>
			</div>
		</div>
	{/if}

	{#if showRelayResetConfirmModal}
		<div class="modal-overlay" on:click={() => showRelayResetConfirmModal = false}>
			<div class="modal-content" on:click|stopPropagation>
				<h3>リレー{resettingRelayType}ブロックの結果をリセットしますか？</h3>
				<p>この操作は元に戻せません。本当にこのブロックの結果をリセットしてよろしいですか？</p>
				<div class="modal-actions">
					<button class="cancel-btn" on:click={() => { showRelayResetConfirmModal = false; resettingRelayType = ''; }}>キャンセル</button>
					<button on:click={handleRelayReset} class="reset-btn">リセット</button>
				</div>
			</div>
		</div>
	{/if}

	{#if showResetConfirmModal}
		<div class="modal-overlay" on:click={() => showResetConfirmModal = false}>
			<div class="modal-content" on:click|stopPropagation>
				<h3>試合をリセットしますか？</h3>
				<p>ID: {resettingMatch.id}, R: {resettingMatch.round}</p>
				<p>{resettingMatch.team1_name || '-'} vs {resettingMatch.team2_name || '-'}</p>
				<p>この操作は元に戻せません。本当にこの試合をリセットしてよろしいですか？</p>
				<div class="modal-actions">
					<button class="cancel-btn" on:click={() => { showResetConfirmModal = false; resettingMatch = null; editingSport = ''; }}>キャンセル</button>
					<button on:click={confirmResetMatch} class="reset-btn">リセット</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	/* ===== モダンUIベースCSS ===== */

body {
  font-family: 'Segoe UI', 'Hiragino Sans', 'Meiryo', sans-serif;
  background: #f7f8fa;
  color: #222;
}

.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 2rem 2rem 1rem 2rem;
  background: #fff;
  border-bottom: 1px solid #e5e7eb;
  box-shadow: 0 2px 8px rgba(0,0,0,0.03);
}
.dashboard-title {
  font-size: 2.2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
}
.dashboard-tabs {
  display: flex;
  gap: 1.2rem;
}
.dashboard-tab-btn {
  background: none;
  border: none;
  font-size: 1.1rem;
  font-weight: 500;
  color: #555;
  padding: 0.5rem 1.2rem;
  border-radius: 6px 6px 0 0;
  transition: background 0.2s, color 0.2s;
  cursor: pointer;
}
.dashboard-tab-btn.active, .dashboard-tab-btn:focus {
  background: #e3e8f0;
  color: #1976d2;
}
.logout-btn {
  background: #e53935;
  color: #fff;
  border: none;
  border-radius: 6px;
  padding: 0.5rem 1.2rem;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}
.logout-btn:hover {
  background: #b71c1c;
}

.sports-buttons {
  display: flex;
  gap: 1rem;
  margin: 2rem 0 1.5rem 0;
  justify-content: center;
}
.sports-btn {
  background: #fff;
  border: 1px solid #d1d5db;
  color: #1976d2;
  font-weight: 600;
  font-size: 1.1rem;
  border-radius: 8px;
  padding: 0.7rem 1.5rem;
  cursor: pointer;
  transition: background 0.2s, border 0.2s, color 0.2s;
}
.sports-btn.active, .sports-btn:focus {
  background: #1976d2;
  color: #fff;
  border-color: #1976d2;
}
.sports-btn:hover {
  background: #e3e8f0;
}

main {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem 1rem 3rem 1rem;
}

/* モーダル */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; width: 100vw; height: 100vh;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal-content {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.13);
  padding: 2.5rem 2rem;
  width: 95vw;
  max-width: 420px;
  text-align: center;
}
.modal-actions {
  display: flex;
  gap: 1.2rem;
  justify-content: center;
  margin-top: 2rem;
}
.modal-btn {
  padding: 0.7rem 1.5rem;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
  transition: background 0.2s;
}
.modal-btn.cancel {
  background: #e0e0e0;
  color: #333;
}
.modal-btn.ok {
  background: #1976d2;
  color: #fff;
}
.modal-btn.danger {
  background: #e53935;
  color: #fff;
}
.modal-btn.cancel:hover { background: #bdbdbd; }
.modal-btn.ok:hover { background: #1565c0; }
.modal-btn.danger:hover { background: #b71c1c; }

/* Bracketry Global Styles */
:global(.bracket-match-team) { background-color: #f0f0f0 !important; border: 1px solid #ccc !important; }
:global(.bracket-match-winner .bracket-match-team) { background-color: #d4edda !important; font-weight: bold; }
:global(.bracket-connector) { border-color: #999 !important; }

/* General */
	.dashboard-container { padding: 2rem; font-family: sans-serif; }
	main { text-align: center; }
	h1 { font-size: 2.5rem; margin-right: auto; }
	h2 { margin-top: 2rem; }
	h3 { margin-bottom: 1rem; }
	button { padding: 0.5rem 1rem; color: white; background-color: #5a5a5a; border: none; border-radius: 4px; cursor: pointer; transition: background-color 0.3s; }
	button.active { background-color: #4285f4; }
	button:hover { opacity: 0.9; }
	.update-btn { 
		background: #4285f4; 
		/* margin-top: 1.5rem;  */
		transition: all 0.2s ease;
	}
	.reset-btn {
		background-color: #d93025 !important;
	}
	.cancel-btn {
		background-color: #f0f0f0;
		color: #333;
	}
	.update-btn:hover:not(:disabled) { 
		background: #3367d6; 
		transform: translateY(-1px);
	}
	.update-btn:disabled { 
		background: #ccc; 
		cursor: not-allowed; 
		opacity: 0.6;
	}
	.update-status { 
		font-size: 0.95rem; 
		margin-left: 1rem; 
		font-weight: 500;
	}
	.update-status.success { color: #43a047; }
	.update-status.error { color: #e53935; }
	.update-status.loading { color: #ff9800; }

	/* Header */
	header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid #ccc; flex-wrap: wrap; }
	.dashboard-tabs { margin: 0 auto; }
	.logout-btn { background-color: #d93025; margin-left: 1rem; }
	.logout-btn:hover { background-color: #c5221b; }

	/* Content Sections */
	.sports-buttons, .tournament-selector, .relay-type-selector, .input-tabs { margin-bottom: 1.5rem; display: flex; align-items: center; justify-content: center; gap: 1rem; flex-wrap: wrap; }
	.bracket-area { margin-top: 2rem; overflow-x: auto; }
	.bracket-wrapper { display: flex; align-items: center; justify-content: center; min-width: 800px; }
	.scores-area { width: 100%; }
	.match-input-area { margin-top: 2rem; display: flex; flex-direction: column; align-items: center; width: 100%; }
	.relay-area { margin-top: 2rem; }
	.relay-info { background: #f8f9fa; border-radius: 8px; padding: 1rem; margin: 1rem 0; max-width: 600px; margin-left: auto; margin-right: auto; }
	.relay-info p { margin: 0.5rem 0; font-size: 0.9rem; color: #666; }

	/* Weather Switcher */
	.weather-switcher { display: flex; align-items: center; gap: 8px; margin-left: 2rem; }
	.switch { position: relative; display: inline-block; width: 50px; height: 24px; }
	.switch input { opacity: 0; width: 0; height: 0; }
	.slider { position: absolute; cursor: pointer; top: 0; left: 0; right: 0; bottom: 0; background-color: #ccc; transition: .4s; border-radius: 24px; }
	.slider:before { position: absolute; content: ""; height: 18px; width: 18px; left: 3px; bottom: 3px; background-color: white; transition: .4s; border-radius: 50%; }
	input:checked + .slider { background-color: #2196F3; }
	input:checked + .slider:before { transform: translateX(26px); }

	/* Bracketry Global Styles */
	:global(.bracket-match-team) { background-color: #f0f0f0 !important; border: 1px solid #ccc !important; }
	:global(.bracket-match-winner .bracket-match-team) { background-color: #d4edda !important; font-weight: bold; }
	:global(.bracket-connector) { border-color: #999 !important; }

	/* Scores Table */
	.scores-container { display: flex; width: 100%; border: 1px solid #ddd; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
	.score-category-column { position: sticky; left: 0; z-index: 1; background-color: #fff; flex-shrink: 0; border-right: 1px solid #ddd; }
	.scores-data-wrapper { display: flex; overflow-x: auto; }
	.score-column { flex: 0 0 120px; display: flex; flex-direction: column; border-left: 1px solid #ddd; }
	.score-header { padding: 10px; text-align: center; font-weight: bold; background-color: #f2f2f2; border-bottom: 1px solid #ddd; }
	.score-cell { padding: 10px; text-align: center; border-bottom: 1px solid #ddd; min-width: 150px; }
	.score-column .score-cell { min-width: auto; }
	.score-cell.odd-row { background-color: #f9f9f9; }

	/* 順位表示のスタイル */
	.rank-cell {
		text-align: center;
		padding: 8px !important;
	}

	.rank-badge {
		display: inline-block;
		padding: 0.3rem 0.6rem;
		border-radius: 12px;
		font-weight: bold;
		color: white;
		font-size: 0.85rem;
		min-width: 40px;
	}

	.rank-badge.rank-1 {
		background: linear-gradient(135deg, #ffd700, #ffb300);
		box-shadow: 0 2px 4px rgba(255, 179, 0, 0.3);
	}

	.rank-badge.rank-2 {
		background: linear-gradient(135deg, #c0c0c0, #a0a0a0);
		box-shadow: 0 2px 4px rgba(160, 160, 160, 0.3);
	}

	.rank-badge.rank-3 {
		background: linear-gradient(135deg, #cd7f32, #b8860b);
		box-shadow: 0 2px 4px rgba(184, 134, 11, 0.3);
	}

	.rank-badge.rank-4,
	.rank-badge.rank-5,
	.rank-badge.rank-6,
	.rank-badge.rank-7,
	.rank-badge.rank-8,
	.rank-badge.rank-9,
	.rank-badge.rank-10,
	.rank-badge.rank-11,
	.rank-badge.rank-12,
	.rank-badge.rank-13,
	.rank-badge.rank-14,
	.rank-badge.rank-15,
	.rank-badge.rank-16 {
		background: linear-gradient(135deg, #6c757d, #5a6268);
		box-shadow: 0 2px 4px rgba(90, 98, 104, 0.3);
	}

	/* Match/Relay Input Cards */
	.tournament-edit-card, .relay-input-card { background: #f8f9fa; border-radius: 12px; box-shadow: 0 2px 8px rgba(66,133,244,0.08); padding: 2rem; margin-bottom: 2rem; width: 100%; max-width: 800px; }
	.match-edit-form { display: flex; flex-wrap: wrap; align-items: center; gap: 1rem; background: #fff; border-radius: 8px; padding: 1.5rem; margin-bottom: 1rem; }
	.match-info { font-size: 0.9rem; color: #333; display: flex; flex-wrap: wrap; gap: 1rem; flex-grow: 1; }
	.match-info.finished { background: #f0f0f0; padding: 1rem; border-radius: 6px; width: 100%; }
	.score-inputs { display: flex; align-items: center; gap: 0.5rem; }
	.score-input { width: 4rem; padding: 0.5rem; font-size: 1.1rem; border: 1px solid #4285f4; border-radius: 4px; text-align: center; }
	
	/* Relay D&D List */
	.dnd-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		max-width: 400px;
		margin: 1.5rem auto;
	}
	.dnd-item {
		display: flex;
		align-items: center;
		padding: 1rem;
		background-color: white;
		border: 1px solid #ddd;
		border-radius: 6px;
		box-shadow: 0 1px 3px rgba(0,0,0,0.05);
		cursor: grab;
		user-select: none;
		transition: all 0.2s ease;
		position: relative;
	}
	.dnd-item:hover { 
		box-shadow: 0 2px 8px rgba(0,0,0,0.1); 
		transform: translateY(-1px);
	}
	.dnd-item:active { 
		cursor: grabbing; 
		background-color: #f0f8ff; 
		transform: scale(1.02);
	}
	.rank-badge { font-weight: bold; font-size: 1rem; color: #fff; background-color: #6c757d; border-radius: 4px; padding: 0.3rem 0.6rem; margin-right: 1rem; }
	.class-name { font-size: 1.1rem; }
	.relay-actions { display: flex; justify-content: center; align-items: center; gap: 1rem; margin-top: 1.5rem; }

	/* Modal */
	.modal-overlay { position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
	.modal-content { background: #fff; border-radius: 8px; padding: 2rem; width: 90%; max-width: 500px; text-align: center; }
	.modal-actions { display: flex; gap: 1rem; justify-content: center; margin-top: 1.5rem; }
	.modal-actions button { background: #ccc; }
	.modal-actions .ok-btn { background: #4285f4; }
	.winner-select-container { margin: 1rem 0; }
	.winner-select-container label { margin-right: 0.5rem; }
	.winner-select-container select { padding: 0.5rem; border-radius: 4px; }

	.nav-button {
		padding: 0.5rem 1rem;
		color: white;
		background-color: #5a5a5a;
		border: none;
		border-radius: 4px;
		cursor: pointer;
		transition: background-color 0.3s;
		text-decoration: none;
		display: inline-block;
	}

	/* Responsive */
	@media (max-width: 900px) {
  .dashboard-header {
    flex-direction: column;
    align-items: flex-start;
    padding: 1.2rem 1rem 0.7rem 1rem;
  }
  .dashboard-title {
    font-size: 1.5rem;
  }
  .dashboard-tabs {
    gap: 0.7rem;
  }
  .sports-buttons {
    flex-direction: column;
    gap: 0.7rem;
  }
  main {
    padding: 1rem 0.5rem 2rem 0.5rem;
  }
  .modal-content {
    padding: 1.2rem 0.5rem;
    max-width: 98vw;
  }
}

@media (max-width: 600px) {
  .dashboard-header {
    padding: 0.7rem 0.3rem 0.5rem 0.3rem;
  }
  .dashboard-title {
    font-size: 1.1rem;
  }
  .dashboard-tabs {
    flex-direction: column;
    gap: 0.3rem;
  }
  .sports-buttons {
    flex-direction: column;
    gap: 0.3rem;
  }
  main {
    padding: 0.5rem 0.1rem 1rem 0.1rem;
  }
  .modal-content {
    padding: 0.7rem 0.2rem;
    max-width: 99vw;
  }
}
</style>
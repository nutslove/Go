document.getElementById('fetch').addEventListener('click', fetchData);

function fetchData() {
    const json_data = {
        iam_user: ["minorun365", "lee-testuser-for-iam", "iamuser3", "iamuser4", "iamuser5"],
        // db_user: ["dbuser1", "dbuser2", "dbuser4", "dbuser5", "dbuser7"],
        os_user: ["T232323", "Z121212", "Z343434", "M565656", "M090909", "M101010", "K232323"]
// simplead_user: ["DnsAdmins","Domain Users","S121212","Administrator","P343434"]
    }
    fetch('http://localhost:8000/', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(json_data) // JavaScriptオブジェクトをJSON文字列に変換
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('ネットワークレスポンスが不正です');
        }
        return response.json();
    })
    .then(data => {
        displayData(data);
    })
    .catch(error => {
        console.error('フェッチエラー:', error);
    });
}

function displayData(data) {
    const resultDiv = document.getElementById('result');
    resultDiv.textContent = JSON.stringify(data, null, 2);
}

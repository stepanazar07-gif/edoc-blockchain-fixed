using System;
using System.Collections.Generic;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class SendDocumentPage : ContentPage
    {
        private readonly ApiService _api = new();
        private readonly MyDocument? _initialDocument;
        private readonly UserInfo? _initialReceiver;
        private List<MyDocument> _documents = new();
        private List<UserInfo> _users = new();

        public SendDocumentPage(MyDocument document)
        {
            _initialDocument = document;
            InitializeComponent();
            FilePicker.SelectedIndexChanged += OnFileSelected;
        }

        public SendDocumentPage(UserInfo receiver)
        {
            _initialReceiver = receiver;
            InitializeComponent();
            FilePicker.SelectedIndexChanged += OnFileSelected;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadDataAsync();
        }

        private async System.Threading.Tasks.Task LoadDataAsync()
        {
            _documents = await _api.GetMyDocumentsAsync() ?? new List<MyDocument>();
            _users = await _api.GetAllUsersAsync() ?? new List<UserInfo>();

            FilePicker.ItemsSource = _documents;
            UserPicker.ItemsSource = _users;

            if (_initialDocument != null)
            {
                var index = _documents.FindIndex(d => d.Id == _initialDocument.Id);
                if (index >= 0) FilePicker.SelectedIndex = index;
            }
            if (_initialReceiver != null)
            {
                var index = _users.FindIndex(u => u.Id == _initialReceiver.Id);
                if (index >= 0) UserPicker.SelectedIndex = index;
            }
            UpdateHashPreview();
        }

        private void OnFileSelected(object? sender, EventArgs e)
        {
            UpdateHashPreview();
        }

        private async void OnSendDocumentClicked(object sender, EventArgs e)
        {
            if (FilePicker.SelectedIndex < 0 || FilePicker.SelectedIndex >= _documents.Count)
            {
                await DisplayAlert("Ошибка", "Выберите файл из хранилища.", "OK");
                return;
            }
            if (UserPicker.SelectedIndex < 0 || UserPicker.SelectedIndex >= _users.Count)
            {
                await DisplayAlert("Ошибка", "Выберите получателя.", "OK");
                return;
            }

            var document = _documents[FilePicker.SelectedIndex];
            var receiver = _users[UserPicker.SelectedIndex];
            var result = await _api.ShareDocumentAsync(document.Id, receiver.Id);
            if (result == null)
            {
                ResultLabel.Text = "Не удалось создать передачу файла.";
                return;
            }

            ResultLabel.Text = $"Передача создана.\nID передачи: {result.TransferId}\nSHA-256 для получателя: {result.FileHash}";
            HistoryStorage.Add(new DocumentRecord
            {
                FileName = document.FileName,
                Status = "Файл отправлен",
                Details = $"Получатель: {receiver.Id}",
                FileHash = result.FileHash,
                CounterpartyId = receiver.Id,
                Timestamp = DateTime.Now
            });

            await DisplayAlert("Файл отправлен", $"Передайте получателю этот SHA-256:\n{result.FileHash}", "OK");
        }

        private void UpdateHashPreview()
        {
            if (FilePicker.SelectedIndex >= 0 && FilePicker.SelectedIndex < _documents.Count)
            {
                FileHashLabel.Text = $"SHA-256: {_documents[FilePicker.SelectedIndex].FileHash}";
            }
            else
            {
                FileHashLabel.Text = "";
            }
        }
    }
}

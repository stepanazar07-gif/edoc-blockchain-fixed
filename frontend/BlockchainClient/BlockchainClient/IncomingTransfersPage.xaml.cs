using System;
using System.Collections.ObjectModel;
using System.IO;
using Microsoft.Maui.Controls;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public partial class IncomingTransfersPage : ContentPage
    {
        private readonly ApiService _api = new();

        public ObservableCollection<IncomingTransfer> Transfers { get; } = new();

        public IncomingTransfersPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadTransfersAsync();
        }

        private async void OnRefreshClicked(object sender, EventArgs e)
        {
            await LoadTransfersAsync();
        }

        private async void OnAcceptClicked(object sender, EventArgs e)
        {
            if (sender is not Button { BindingContext: IncomingTransfer transfer }) return;

            var hash = await DisplayPromptAsync(
                "Проверка SHA-256",
                "Введите хеш-сумму, которую сообщил отправитель.",
                "Принять",
                "Отмена",
                keyboard: Keyboard.Text);
            if (string.IsNullOrWhiteSpace(hash)) return;

            var accepted = await _api.AcceptTransferAsync(transfer.TransferId, hash.Trim());
            if (accepted == null)
            {
                await DisplayAlert("Доступ запрещён", "Хеш не совпал или передача уже обработана.", "OK");
                return;
            }

            var fileBytes = await _api.DownloadFileAsync(accepted.FileHash);
            if (fileBytes == null)
            {
                await DisplayAlert("Ошибка", "Файл принят, но скачать его не удалось.", "OK");
                return;
            }

            var localPath = Path.Combine(FileSystem.Current.AppDataDirectory, accepted.FileName);
            await File.WriteAllBytesAsync(localPath, fileBytes);

            HistoryStorage.Add(new DocumentRecord
            {
                FileName = accepted.FileName,
                Status = "Файл принят",
                Details = $"Отправитель: {accepted.SenderId}",
                FileHash = accepted.FileHash,
                CounterpartyId = accepted.SenderId,
                Timestamp = DateTime.Now
            });

            await DisplayAlert("Готово", $"Файл принят и сохранён:\n{localPath}", "OK");
            await Launcher.OpenAsync(new OpenFileRequest { File = new ReadOnlyFile(localPath) });
            await LoadTransfersAsync();
        }

        private async void OnDeclineClicked(object sender, EventArgs e)
        {
            if (sender is not Button { BindingContext: IncomingTransfer transfer }) return;

            var ok = await _api.DeclineTransferAsync(transfer.TransferId);
            if (!ok)
            {
                await DisplayAlert("Ошибка", "Не удалось отказаться от файла.", "OK");
                return;
            }

            HistoryStorage.Add(new DocumentRecord
            {
                FileName = transfer.FileName,
                Status = "Файл не принят",
                Details = $"Отправитель: {transfer.SenderId}",
                CounterpartyId = transfer.SenderId,
                Timestamp = DateTime.Now
            });

            await LoadTransfersAsync();
        }

        private async System.Threading.Tasks.Task LoadTransfersAsync()
        {
            var list = await _api.GetIncomingTransfersAsync();
            Transfers.Clear();
            if (list == null) return;
            foreach (var transfer in list)
            {
                Transfers.Add(transfer);
            }
        }
    }
}

using System;
using System.Collections.ObjectModel;
using System.IO;
using Microsoft.Maui.Controls;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public partial class UploadPage : ContentPage
    {
        private readonly ApiService _api = new();
        private byte[]? _fileBytes;
        private string _fileName = "";
        private string _mimeType = "application/octet-stream";

        public ObservableCollection<MyDocument> UploadedDocuments { get; } = new();

        public UploadPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        private async void OnSelectFileClicked(object sender, EventArgs e)
        {
            try
            {
                var result = await FilePicker.PickAsync(new PickOptions
                {
                    PickerTitle = "Выберите файл"
                });
                if (result == null) return;

                await using var stream = await result.OpenReadAsync();
                using var memory = new MemoryStream();
                await stream.CopyToAsync(memory);

                _fileBytes = memory.ToArray();
                _fileName = result.FileName;
                _mimeType = ApiService.GuessMimeType(_fileName);

                var hash = ApiService.ComputeSha256(_fileBytes);
                FileInfoLabel.Text = $"{_fileName} · {FileSizeFormatter.FormatBytes(_fileBytes.Length)}";
                HashPreviewLabel.Text = $"SHA-256: {hash}";
                UploadButton.IsEnabled = true;
            }
            catch (Exception ex)
            {
                await DisplayAlert("Ошибка", ex.Message, "OK");
            }
        }

        private async void OnUploadClicked(object sender, EventArgs e)
        {
            if (_fileBytes == null || string.IsNullOrWhiteSpace(_fileName)) return;

            UploadIndicator.IsRunning = true;
            UploadIndicator.IsVisible = true;
            UploadButton.IsEnabled = false;

            try
            {
                var result = await _api.UploadDocumentAsync(_fileName, _fileBytes, _mimeType);
                if (result == null)
                {
                    var errorText = string.IsNullOrWhiteSpace(_api.LastError)
                        ? "Сервер не принял файл."
                        : _api.LastError;
                    await DisplayAlert("Ошибка", errorText, "OK");
                    return;
                }

                var now = DateTime.Now;
                UploadedDocuments.Insert(0, new MyDocument
                {
                    Id = string.IsNullOrWhiteSpace(result.FileId) ? result.DocumentId : result.FileId,
                    FileName = result.FileName,
                    FileHash = result.FileHash,
                    FileSize = result.FileSize,
                    MimeType = result.MimeType,
                    UploadDate = now.ToString("yyyy-MM-dd"),
                    UploadTime = now.ToString("HH:mm:ss"),
                    UploadedBy = result.UploadedBy
                });

                HistoryStorage.Add(new DocumentRecord
                {
                    FileName = result.FileName,
                    Status = "Файл загружен",
                    Details = $"Блок: {result.BlockHeight}",
                    FileHash = result.FileHash,
                    Timestamp = DateTime.Now
                });

                AppMessenger.NotifyFileUploaded();
                await DisplayAlert("Готово", $"Файл загружен.\nSHA-256: {result.FileHash}", "OK");

                _fileBytes = null;
                _fileName = "";
                FileInfoLabel.Text = "Файл не выбран";
                HashPreviewLabel.Text = "";
            }
            catch (Exception ex)
            {
                await DisplayAlert("Ошибка", ex.Message, "OK");
            }
            finally
            {
                UploadIndicator.IsRunning = false;
                UploadIndicator.IsVisible = false;
                UploadButton.IsEnabled = _fileBytes != null;
            }
        }
    }
}

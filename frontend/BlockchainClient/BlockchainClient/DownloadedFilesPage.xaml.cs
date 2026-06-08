using System;
using System.Collections.ObjectModel;
using System.IO;
using Microsoft.Maui.Controls;
using Microsoft.Maui.Storage;

namespace BlockchainClient
{
    public partial class DownloadedFilesPage : ContentPage
    {
        private readonly ApiService _api = new();

        public ObservableCollection<ReceivedFile> Files { get; } = new();

        public DownloadedFilesPage()
        {
            InitializeComponent();
            BindingContext = this;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadFilesAsync();
        }

        private async void OnRefreshClicked(object sender, EventArgs e)
        {
            await LoadFilesAsync();
        }

        private async void OnOpenClicked(object sender, EventArgs e)
        {
            if (sender is not Button { BindingContext: ReceivedFile file }) return;

            var bytes = await _api.DownloadFileAsync(file.FileHash);
            if (bytes == null)
            {
                await DisplayAlert("Ошибка", "Не удалось скачать файл.", "OK");
                return;
            }

            var localPath = Path.Combine(FileSystem.Current.AppDataDirectory, file.FileName);
            await File.WriteAllBytesAsync(localPath, bytes);
            await Launcher.OpenAsync(new OpenFileRequest { File = new ReadOnlyFile(localPath) });
        }

        private async System.Threading.Tasks.Task LoadFilesAsync()
        {
            var list = await _api.GetReceivedFilesAsync();
            Files.Clear();
            if (list == null) return;
            foreach (var file in list)
            {
                Files.Add(file);
            }
        }
    }
}

using System;
using System.Collections.ObjectModel;
using Microsoft.Maui.ApplicationModel;
using Microsoft.Maui.Controls;

namespace BlockchainClient
{
    public partial class MyDocumentsPage : ContentPage
    {
        private readonly ApiService _api = new();

        public ObservableCollection<MyDocument> Documents { get; } = new();

        public MyDocumentsPage()
        {
            InitializeComponent();
            BindingContext = this;
            AppMessenger.OnFileUploaded += OnFileUploaded;
        }

        protected override async void OnAppearing()
        {
            base.OnAppearing();
            await LoadDocumentsAsync();
        }

        protected override void OnDisappearing()
        {
            base.OnDisappearing();
            AppMessenger.OnFileUploaded -= OnFileUploaded;
        }

        private async void OnFileUploaded()
        {
            await MainThread.InvokeOnMainThreadAsync(LoadDocumentsAsync);
        }

        private async void OnRefreshClicked(object sender, EventArgs e)
        {
            await LoadDocumentsAsync();
        }

        private async void OnUploadClicked(object sender, EventArgs e)
        {
            await Shell.Current.GoToAsync("//UploadPage");
        }

        private async void OnSendClicked(object sender, EventArgs e)
        {
            if (sender is Button { BindingContext: MyDocument doc })
            {
                await Navigation.PushAsync(new SendDocumentPage(doc));
            }
        }

        private async System.Threading.Tasks.Task LoadDocumentsAsync()
        {
            var docs = await _api.GetMyDocumentsAsync();
            if (docs == null) return;

            Documents.Clear();
            foreach (var doc in docs)
            {
                Documents.Add(doc);
            }
        }
    }
}

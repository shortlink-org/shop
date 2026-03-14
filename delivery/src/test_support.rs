#[cfg(test)]
use std::ops::{Deref, DerefMut};

#[cfg(test)]
use testcontainers::{ContainerAsync, Image};

/// Test-only wrapper that keeps testcontainers deterministic in environments
/// where auto-remove mode is unstable and performs explicit cleanup on drop.
#[cfg(test)]
pub struct ManagedAsyncContainer<I: Image + Send + 'static> {
    inner: Option<ContainerAsync<I>>,
}

#[cfg(test)]
impl<I: Image + Send + 'static> ManagedAsyncContainer<I> {
    pub fn new(inner: ContainerAsync<I>) -> Self {
        Self { inner: Some(inner) }
    }

    pub async fn rm(mut self) -> testcontainers::core::error::Result<()> {
        if let Some(container) = self.inner.take() {
            container.rm().await
        } else {
            Ok(())
        }
    }
}

#[cfg(test)]
impl<I: Image + Send + 'static> Deref for ManagedAsyncContainer<I> {
    type Target = ContainerAsync<I>;

    fn deref(&self) -> &Self::Target {
        self.inner.as_ref().expect("container already removed")
    }
}

#[cfg(test)]
impl<I: Image + Send + 'static> DerefMut for ManagedAsyncContainer<I> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        self.inner.as_mut().expect("container already removed")
    }
}

#[cfg(test)]
impl<I: Image + Send + 'static> Drop for ManagedAsyncContainer<I> {
    fn drop(&mut self) {
        let Some(container) = self.inner.take() else {
            return;
        };

        let _ = std::thread::spawn(move || {
            let Ok(runtime) = tokio::runtime::Builder::new_current_thread()
                .enable_all()
                .build()
            else {
                return;
            };

            let _ = runtime.block_on(async move { container.rm().await });
        })
        .join();
    }
}

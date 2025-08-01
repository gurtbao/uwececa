// Implement From<A> for B where B implements Into<C> and a Implements From<C>.
#[macro_export]
macro_rules! transitive_from {
    ($from:ty, $target:ty, $intermediate:ty) => {
        impl From<$from> for $target {
            fn from(value: $from) -> Self {
                <$target>::from(<$intermediate>::from(value))
            }
        }
    };
}
